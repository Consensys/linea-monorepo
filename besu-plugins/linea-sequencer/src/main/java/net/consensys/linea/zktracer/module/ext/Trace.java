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

package net.consensys.linea.zktracer.module.ext;

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
    @JsonProperty("ACC_DELTA_0") List<BigInteger> accDelta0,
    @JsonProperty("ACC_DELTA_1") List<BigInteger> accDelta1,
    @JsonProperty("ACC_DELTA_2") List<BigInteger> accDelta2,
    @JsonProperty("ACC_DELTA_3") List<BigInteger> accDelta3,
    @JsonProperty("ACC_H_0") List<BigInteger> accH0,
    @JsonProperty("ACC_H_1") List<BigInteger> accH1,
    @JsonProperty("ACC_H_2") List<BigInteger> accH2,
    @JsonProperty("ACC_H_3") List<BigInteger> accH3,
    @JsonProperty("ACC_H_4") List<BigInteger> accH4,
    @JsonProperty("ACC_H_5") List<BigInteger> accH5,
    @JsonProperty("ACC_I_0") List<BigInteger> accI0,
    @JsonProperty("ACC_I_1") List<BigInteger> accI1,
    @JsonProperty("ACC_I_2") List<BigInteger> accI2,
    @JsonProperty("ACC_I_3") List<BigInteger> accI3,
    @JsonProperty("ACC_I_4") List<BigInteger> accI4,
    @JsonProperty("ACC_I_5") List<BigInteger> accI5,
    @JsonProperty("ACC_I_6") List<BigInteger> accI6,
    @JsonProperty("ACC_J_0") List<BigInteger> accJ0,
    @JsonProperty("ACC_J_1") List<BigInteger> accJ1,
    @JsonProperty("ACC_J_2") List<BigInteger> accJ2,
    @JsonProperty("ACC_J_3") List<BigInteger> accJ3,
    @JsonProperty("ACC_J_4") List<BigInteger> accJ4,
    @JsonProperty("ACC_J_5") List<BigInteger> accJ5,
    @JsonProperty("ACC_J_6") List<BigInteger> accJ6,
    @JsonProperty("ACC_J_7") List<BigInteger> accJ7,
    @JsonProperty("ACC_Q_0") List<BigInteger> accQ0,
    @JsonProperty("ACC_Q_1") List<BigInteger> accQ1,
    @JsonProperty("ACC_Q_2") List<BigInteger> accQ2,
    @JsonProperty("ACC_Q_3") List<BigInteger> accQ3,
    @JsonProperty("ACC_Q_4") List<BigInteger> accQ4,
    @JsonProperty("ACC_Q_5") List<BigInteger> accQ5,
    @JsonProperty("ACC_Q_6") List<BigInteger> accQ6,
    @JsonProperty("ACC_Q_7") List<BigInteger> accQ7,
    @JsonProperty("ACC_R_0") List<BigInteger> accR0,
    @JsonProperty("ACC_R_1") List<BigInteger> accR1,
    @JsonProperty("ACC_R_2") List<BigInteger> accR2,
    @JsonProperty("ACC_R_3") List<BigInteger> accR3,
    @JsonProperty("ARG_1_HI") List<BigInteger> arg1Hi,
    @JsonProperty("ARG_1_LO") List<BigInteger> arg1Lo,
    @JsonProperty("ARG_2_HI") List<BigInteger> arg2Hi,
    @JsonProperty("ARG_2_LO") List<BigInteger> arg2Lo,
    @JsonProperty("ARG_3_HI") List<BigInteger> arg3Hi,
    @JsonProperty("ARG_3_LO") List<BigInteger> arg3Lo,
    @JsonProperty("BIT_1") List<Boolean> bit1,
    @JsonProperty("BIT_2") List<Boolean> bit2,
    @JsonProperty("BIT_3") List<Boolean> bit3,
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
    @JsonProperty("BYTE_DELTA_0") List<UnsignedByte> byteDelta0,
    @JsonProperty("BYTE_DELTA_1") List<UnsignedByte> byteDelta1,
    @JsonProperty("BYTE_DELTA_2") List<UnsignedByte> byteDelta2,
    @JsonProperty("BYTE_DELTA_3") List<UnsignedByte> byteDelta3,
    @JsonProperty("BYTE_H_0") List<UnsignedByte> byteH0,
    @JsonProperty("BYTE_H_1") List<UnsignedByte> byteH1,
    @JsonProperty("BYTE_H_2") List<UnsignedByte> byteH2,
    @JsonProperty("BYTE_H_3") List<UnsignedByte> byteH3,
    @JsonProperty("BYTE_H_4") List<UnsignedByte> byteH4,
    @JsonProperty("BYTE_H_5") List<UnsignedByte> byteH5,
    @JsonProperty("BYTE_I_0") List<UnsignedByte> byteI0,
    @JsonProperty("BYTE_I_1") List<UnsignedByte> byteI1,
    @JsonProperty("BYTE_I_2") List<UnsignedByte> byteI2,
    @JsonProperty("BYTE_I_3") List<UnsignedByte> byteI3,
    @JsonProperty("BYTE_I_4") List<UnsignedByte> byteI4,
    @JsonProperty("BYTE_I_5") List<UnsignedByte> byteI5,
    @JsonProperty("BYTE_I_6") List<UnsignedByte> byteI6,
    @JsonProperty("BYTE_J_0") List<UnsignedByte> byteJ0,
    @JsonProperty("BYTE_J_1") List<UnsignedByte> byteJ1,
    @JsonProperty("BYTE_J_2") List<UnsignedByte> byteJ2,
    @JsonProperty("BYTE_J_3") List<UnsignedByte> byteJ3,
    @JsonProperty("BYTE_J_4") List<UnsignedByte> byteJ4,
    @JsonProperty("BYTE_J_5") List<UnsignedByte> byteJ5,
    @JsonProperty("BYTE_J_6") List<UnsignedByte> byteJ6,
    @JsonProperty("BYTE_J_7") List<UnsignedByte> byteJ7,
    @JsonProperty("BYTE_Q_0") List<UnsignedByte> byteQ0,
    @JsonProperty("BYTE_Q_1") List<UnsignedByte> byteQ1,
    @JsonProperty("BYTE_Q_2") List<UnsignedByte> byteQ2,
    @JsonProperty("BYTE_Q_3") List<UnsignedByte> byteQ3,
    @JsonProperty("BYTE_Q_4") List<UnsignedByte> byteQ4,
    @JsonProperty("BYTE_Q_5") List<UnsignedByte> byteQ5,
    @JsonProperty("BYTE_Q_6") List<UnsignedByte> byteQ6,
    @JsonProperty("BYTE_Q_7") List<UnsignedByte> byteQ7,
    @JsonProperty("BYTE_R_0") List<UnsignedByte> byteR0,
    @JsonProperty("BYTE_R_1") List<UnsignedByte> byteR1,
    @JsonProperty("BYTE_R_2") List<UnsignedByte> byteR2,
    @JsonProperty("BYTE_R_3") List<UnsignedByte> byteR3,
    @JsonProperty("CMP") List<Boolean> cmp,
    @JsonProperty("CT") List<BigInteger> ct,
    @JsonProperty("INST") List<BigInteger> inst,
    @JsonProperty("OF_H") List<Boolean> ofH,
    @JsonProperty("OF_I") List<Boolean> ofI,
    @JsonProperty("OF_J") List<Boolean> ofJ,
    @JsonProperty("OF_RES") List<Boolean> ofRes,
    @JsonProperty("OLI") List<Boolean> oli,
    @JsonProperty("RES_HI") List<BigInteger> resHi,
    @JsonProperty("RES_LO") List<BigInteger> resLo,
    @JsonProperty("STAMP") List<BigInteger> stamp) {
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
    private final List<BigInteger> accDelta0 = new ArrayList<>();
    private final List<BigInteger> accDelta1 = new ArrayList<>();
    private final List<BigInteger> accDelta2 = new ArrayList<>();
    private final List<BigInteger> accDelta3 = new ArrayList<>();
    private final List<BigInteger> accH0 = new ArrayList<>();
    private final List<BigInteger> accH1 = new ArrayList<>();
    private final List<BigInteger> accH2 = new ArrayList<>();
    private final List<BigInteger> accH3 = new ArrayList<>();
    private final List<BigInteger> accH4 = new ArrayList<>();
    private final List<BigInteger> accH5 = new ArrayList<>();
    private final List<BigInteger> accI0 = new ArrayList<>();
    private final List<BigInteger> accI1 = new ArrayList<>();
    private final List<BigInteger> accI2 = new ArrayList<>();
    private final List<BigInteger> accI3 = new ArrayList<>();
    private final List<BigInteger> accI4 = new ArrayList<>();
    private final List<BigInteger> accI5 = new ArrayList<>();
    private final List<BigInteger> accI6 = new ArrayList<>();
    private final List<BigInteger> accJ0 = new ArrayList<>();
    private final List<BigInteger> accJ1 = new ArrayList<>();
    private final List<BigInteger> accJ2 = new ArrayList<>();
    private final List<BigInteger> accJ3 = new ArrayList<>();
    private final List<BigInteger> accJ4 = new ArrayList<>();
    private final List<BigInteger> accJ5 = new ArrayList<>();
    private final List<BigInteger> accJ6 = new ArrayList<>();
    private final List<BigInteger> accJ7 = new ArrayList<>();
    private final List<BigInteger> accQ0 = new ArrayList<>();
    private final List<BigInteger> accQ1 = new ArrayList<>();
    private final List<BigInteger> accQ2 = new ArrayList<>();
    private final List<BigInteger> accQ3 = new ArrayList<>();
    private final List<BigInteger> accQ4 = new ArrayList<>();
    private final List<BigInteger> accQ5 = new ArrayList<>();
    private final List<BigInteger> accQ6 = new ArrayList<>();
    private final List<BigInteger> accQ7 = new ArrayList<>();
    private final List<BigInteger> accR0 = new ArrayList<>();
    private final List<BigInteger> accR1 = new ArrayList<>();
    private final List<BigInteger> accR2 = new ArrayList<>();
    private final List<BigInteger> accR3 = new ArrayList<>();
    private final List<BigInteger> arg1Hi = new ArrayList<>();
    private final List<BigInteger> arg1Lo = new ArrayList<>();
    private final List<BigInteger> arg2Hi = new ArrayList<>();
    private final List<BigInteger> arg2Lo = new ArrayList<>();
    private final List<BigInteger> arg3Hi = new ArrayList<>();
    private final List<BigInteger> arg3Lo = new ArrayList<>();
    private final List<Boolean> bit1 = new ArrayList<>();
    private final List<Boolean> bit2 = new ArrayList<>();
    private final List<Boolean> bit3 = new ArrayList<>();
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
    private final List<UnsignedByte> byteDelta0 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta1 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta2 = new ArrayList<>();
    private final List<UnsignedByte> byteDelta3 = new ArrayList<>();
    private final List<UnsignedByte> byteH0 = new ArrayList<>();
    private final List<UnsignedByte> byteH1 = new ArrayList<>();
    private final List<UnsignedByte> byteH2 = new ArrayList<>();
    private final List<UnsignedByte> byteH3 = new ArrayList<>();
    private final List<UnsignedByte> byteH4 = new ArrayList<>();
    private final List<UnsignedByte> byteH5 = new ArrayList<>();
    private final List<UnsignedByte> byteI0 = new ArrayList<>();
    private final List<UnsignedByte> byteI1 = new ArrayList<>();
    private final List<UnsignedByte> byteI2 = new ArrayList<>();
    private final List<UnsignedByte> byteI3 = new ArrayList<>();
    private final List<UnsignedByte> byteI4 = new ArrayList<>();
    private final List<UnsignedByte> byteI5 = new ArrayList<>();
    private final List<UnsignedByte> byteI6 = new ArrayList<>();
    private final List<UnsignedByte> byteJ0 = new ArrayList<>();
    private final List<UnsignedByte> byteJ1 = new ArrayList<>();
    private final List<UnsignedByte> byteJ2 = new ArrayList<>();
    private final List<UnsignedByte> byteJ3 = new ArrayList<>();
    private final List<UnsignedByte> byteJ4 = new ArrayList<>();
    private final List<UnsignedByte> byteJ5 = new ArrayList<>();
    private final List<UnsignedByte> byteJ6 = new ArrayList<>();
    private final List<UnsignedByte> byteJ7 = new ArrayList<>();
    private final List<UnsignedByte> byteQ0 = new ArrayList<>();
    private final List<UnsignedByte> byteQ1 = new ArrayList<>();
    private final List<UnsignedByte> byteQ2 = new ArrayList<>();
    private final List<UnsignedByte> byteQ3 = new ArrayList<>();
    private final List<UnsignedByte> byteQ4 = new ArrayList<>();
    private final List<UnsignedByte> byteQ5 = new ArrayList<>();
    private final List<UnsignedByte> byteQ6 = new ArrayList<>();
    private final List<UnsignedByte> byteQ7 = new ArrayList<>();
    private final List<UnsignedByte> byteR0 = new ArrayList<>();
    private final List<UnsignedByte> byteR1 = new ArrayList<>();
    private final List<UnsignedByte> byteR2 = new ArrayList<>();
    private final List<UnsignedByte> byteR3 = new ArrayList<>();
    private final List<Boolean> cmp = new ArrayList<>();
    private final List<BigInteger> ct = new ArrayList<>();
    private final List<BigInteger> inst = new ArrayList<>();
    private final List<Boolean> ofH = new ArrayList<>();
    private final List<Boolean> ofI = new ArrayList<>();
    private final List<Boolean> ofJ = new ArrayList<>();
    private final List<Boolean> ofRes = new ArrayList<>();
    private final List<Boolean> oli = new ArrayList<>();
    private final List<BigInteger> resHi = new ArrayList<>();
    private final List<BigInteger> resLo = new ArrayList<>();
    private final List<BigInteger> stamp = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder accA0(final BigInteger b) {
      if (filled.get(66)) {
        throw new IllegalStateException("ACC_A_0 already set");
      } else {
        filled.set(66);
      }

      accA0.add(b);

      return this;
    }

    TraceBuilder accA1(final BigInteger b) {
      if (filled.get(86)) {
        throw new IllegalStateException("ACC_A_1 already set");
      } else {
        filled.set(86);
      }

      accA1.add(b);

      return this;
    }

    TraceBuilder accA2(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ACC_A_2 already set");
      } else {
        filled.set(6);
      }

      accA2.add(b);

      return this;
    }

    TraceBuilder accA3(final BigInteger b) {
      if (filled.get(79)) {
        throw new IllegalStateException("ACC_A_3 already set");
      } else {
        filled.set(79);
      }

      accA3.add(b);

      return this;
    }

    TraceBuilder accB0(final BigInteger b) {
      if (filled.get(70)) {
        throw new IllegalStateException("ACC_B_0 already set");
      } else {
        filled.set(70);
      }

      accB0.add(b);

      return this;
    }

    TraceBuilder accB1(final BigInteger b) {
      if (filled.get(34)) {
        throw new IllegalStateException("ACC_B_1 already set");
      } else {
        filled.set(34);
      }

      accB1.add(b);

      return this;
    }

    TraceBuilder accB2(final BigInteger b) {
      if (filled.get(98)) {
        throw new IllegalStateException("ACC_B_2 already set");
      } else {
        filled.set(98);
      }

      accB2.add(b);

      return this;
    }

    TraceBuilder accB3(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("ACC_B_3 already set");
      } else {
        filled.set(54);
      }

      accB3.add(b);

      return this;
    }

    TraceBuilder accC0(final BigInteger b) {
      if (filled.get(83)) {
        throw new IllegalStateException("ACC_C_0 already set");
      } else {
        filled.set(83);
      }

      accC0.add(b);

      return this;
    }

    TraceBuilder accC1(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("ACC_C_1 already set");
      } else {
        filled.set(46);
      }

      accC1.add(b);

      return this;
    }

    TraceBuilder accC2(final BigInteger b) {
      if (filled.get(62)) {
        throw new IllegalStateException("ACC_C_2 already set");
      } else {
        filled.set(62);
      }

      accC2.add(b);

      return this;
    }

    TraceBuilder accC3(final BigInteger b) {
      if (filled.get(114)) {
        throw new IllegalStateException("ACC_C_3 already set");
      } else {
        filled.set(114);
      }

      accC3.add(b);

      return this;
    }

    TraceBuilder accDelta0(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("ACC_DELTA_0 already set");
      } else {
        filled.set(50);
      }

      accDelta0.add(b);

      return this;
    }

    TraceBuilder accDelta1(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("ACC_DELTA_1 already set");
      } else {
        filled.set(30);
      }

      accDelta1.add(b);

      return this;
    }

    TraceBuilder accDelta2(final BigInteger b) {
      if (filled.get(113)) {
        throw new IllegalStateException("ACC_DELTA_2 already set");
      } else {
        filled.set(113);
      }

      accDelta2.add(b);

      return this;
    }

    TraceBuilder accDelta3(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("ACC_DELTA_3 already set");
      } else {
        filled.set(11);
      }

      accDelta3.add(b);

      return this;
    }

    TraceBuilder accH0(final BigInteger b) {
      if (filled.get(95)) {
        throw new IllegalStateException("ACC_H_0 already set");
      } else {
        filled.set(95);
      }

      accH0.add(b);

      return this;
    }

    TraceBuilder accH1(final BigInteger b) {
      if (filled.get(67)) {
        throw new IllegalStateException("ACC_H_1 already set");
      } else {
        filled.set(67);
      }

      accH1.add(b);

      return this;
    }

    TraceBuilder accH2(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("ACC_H_2 already set");
      } else {
        filled.set(36);
      }

      accH2.add(b);

      return this;
    }

    TraceBuilder accH3(final BigInteger b) {
      if (filled.get(76)) {
        throw new IllegalStateException("ACC_H_3 already set");
      } else {
        filled.set(76);
      }

      accH3.add(b);

      return this;
    }

    TraceBuilder accH4(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("ACC_H_4 already set");
      } else {
        filled.set(16);
      }

      accH4.add(b);

      return this;
    }

    TraceBuilder accH5(final BigInteger b) {
      if (filled.get(81)) {
        throw new IllegalStateException("ACC_H_5 already set");
      } else {
        filled.set(81);
      }

      accH5.add(b);

      return this;
    }

    TraceBuilder accI0(final BigInteger b) {
      if (filled.get(68)) {
        throw new IllegalStateException("ACC_I_0 already set");
      } else {
        filled.set(68);
      }

      accI0.add(b);

      return this;
    }

    TraceBuilder accI1(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("ACC_I_1 already set");
      } else {
        filled.set(51);
      }

      accI1.add(b);

      return this;
    }

    TraceBuilder accI2(final BigInteger b) {
      if (filled.get(105)) {
        throw new IllegalStateException("ACC_I_2 already set");
      } else {
        filled.set(105);
      }

      accI2.add(b);

      return this;
    }

    TraceBuilder accI3(final BigInteger b) {
      if (filled.get(77)) {
        throw new IllegalStateException("ACC_I_3 already set");
      } else {
        filled.set(77);
      }

      accI3.add(b);

      return this;
    }

    TraceBuilder accI4(final BigInteger b) {
      if (filled.get(73)) {
        throw new IllegalStateException("ACC_I_4 already set");
      } else {
        filled.set(73);
      }

      accI4.add(b);

      return this;
    }

    TraceBuilder accI5(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("ACC_I_5 already set");
      } else {
        filled.set(17);
      }

      accI5.add(b);

      return this;
    }

    TraceBuilder accI6(final BigInteger b) {
      if (filled.get(69)) {
        throw new IllegalStateException("ACC_I_6 already set");
      } else {
        filled.set(69);
      }

      accI6.add(b);

      return this;
    }

    TraceBuilder accJ0(final BigInteger b) {
      if (filled.get(93)) {
        throw new IllegalStateException("ACC_J_0 already set");
      } else {
        filled.set(93);
      }

      accJ0.add(b);

      return this;
    }

    TraceBuilder accJ1(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ACC_J_1 already set");
      } else {
        filled.set(25);
      }

      accJ1.add(b);

      return this;
    }

    TraceBuilder accJ2(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("ACC_J_2 already set");
      } else {
        filled.set(57);
      }

      accJ2.add(b);

      return this;
    }

    TraceBuilder accJ3(final BigInteger b) {
      if (filled.get(64)) {
        throw new IllegalStateException("ACC_J_3 already set");
      } else {
        filled.set(64);
      }

      accJ3.add(b);

      return this;
    }

    TraceBuilder accJ4(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("ACC_J_4 already set");
      } else {
        filled.set(56);
      }

      accJ4.add(b);

      return this;
    }

    TraceBuilder accJ5(final BigInteger b) {
      if (filled.get(85)) {
        throw new IllegalStateException("ACC_J_5 already set");
      } else {
        filled.set(85);
      }

      accJ5.add(b);

      return this;
    }

    TraceBuilder accJ6(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("ACC_J_6 already set");
      } else {
        filled.set(26);
      }

      accJ6.add(b);

      return this;
    }

    TraceBuilder accJ7(final BigInteger b) {
      if (filled.get(89)) {
        throw new IllegalStateException("ACC_J_7 already set");
      } else {
        filled.set(89);
      }

      accJ7.add(b);

      return this;
    }

    TraceBuilder accQ0(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("ACC_Q_0 already set");
      } else {
        filled.set(44);
      }

      accQ0.add(b);

      return this;
    }

    TraceBuilder accQ1(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_Q_1 already set");
      } else {
        filled.set(3);
      }

      accQ1.add(b);

      return this;
    }

    TraceBuilder accQ2(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("ACC_Q_2 already set");
      } else {
        filled.set(29);
      }

      accQ2.add(b);

      return this;
    }

    TraceBuilder accQ3(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_Q_3 already set");
      } else {
        filled.set(1);
      }

      accQ3.add(b);

      return this;
    }

    TraceBuilder accQ4(final BigInteger b) {
      if (filled.get(94)) {
        throw new IllegalStateException("ACC_Q_4 already set");
      } else {
        filled.set(94);
      }

      accQ4.add(b);

      return this;
    }

    TraceBuilder accQ5(final BigInteger b) {
      if (filled.get(116)) {
        throw new IllegalStateException("ACC_Q_5 already set");
      } else {
        filled.set(116);
      }

      accQ5.add(b);

      return this;
    }

    TraceBuilder accQ6(final BigInteger b) {
      if (filled.get(32)) {
        throw new IllegalStateException("ACC_Q_6 already set");
      } else {
        filled.set(32);
      }

      accQ6.add(b);

      return this;
    }

    TraceBuilder accQ7(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("ACC_Q_7 already set");
      } else {
        filled.set(24);
      }

      accQ7.add(b);

      return this;
    }

    TraceBuilder accR0(final BigInteger b) {
      if (filled.get(104)) {
        throw new IllegalStateException("ACC_R_0 already set");
      } else {
        filled.set(104);
      }

      accR0.add(b);

      return this;
    }

    TraceBuilder accR1(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("ACC_R_1 already set");
      } else {
        filled.set(14);
      }

      accR1.add(b);

      return this;
    }

    TraceBuilder accR2(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("ACC_R_2 already set");
      } else {
        filled.set(49);
      }

      accR2.add(b);

      return this;
    }

    TraceBuilder accR3(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("ACC_R_3 already set");
      } else {
        filled.set(33);
      }

      accR3.add(b);

      return this;
    }

    TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(38);
      }

      arg1Hi.add(b);

      return this;
    }

    TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(15);
      }

      arg1Lo.add(b);

      return this;
    }

    TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(115);
      }

      arg2Hi.add(b);

      return this;
    }

    TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(53);
      }

      arg2Lo.add(b);

      return this;
    }

    TraceBuilder arg3Hi(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("ARG_3_HI already set");
      } else {
        filled.set(40);
      }

      arg3Hi.add(b);

      return this;
    }

    TraceBuilder arg3Lo(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ARG_3_LO already set");
      } else {
        filled.set(8);
      }

      arg3Lo.add(b);

      return this;
    }

    TraceBuilder bit1(final Boolean b) {
      if (filled.get(101)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(101);
      }

      bit1.add(b);

      return this;
    }

    TraceBuilder bit2(final Boolean b) {
      if (filled.get(97)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(97);
      }

      bit2.add(b);

      return this;
    }

    TraceBuilder bit3(final Boolean b) {
      if (filled.get(4)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(4);
      }

      bit3.add(b);

      return this;
    }

    TraceBuilder byteA0(final UnsignedByte b) {
      if (filled.get(91)) {
        throw new IllegalStateException("BYTE_A_0 already set");
      } else {
        filled.set(91);
      }

      byteA0.add(b);

      return this;
    }

    TraceBuilder byteA1(final UnsignedByte b) {
      if (filled.get(90)) {
        throw new IllegalStateException("BYTE_A_1 already set");
      } else {
        filled.set(90);
      }

      byteA1.add(b);

      return this;
    }

    TraceBuilder byteA2(final UnsignedByte b) {
      if (filled.get(102)) {
        throw new IllegalStateException("BYTE_A_2 already set");
      } else {
        filled.set(102);
      }

      byteA2.add(b);

      return this;
    }

    TraceBuilder byteA3(final UnsignedByte b) {
      if (filled.get(48)) {
        throw new IllegalStateException("BYTE_A_3 already set");
      } else {
        filled.set(48);
      }

      byteA3.add(b);

      return this;
    }

    TraceBuilder byteB0(final UnsignedByte b) {
      if (filled.get(28)) {
        throw new IllegalStateException("BYTE_B_0 already set");
      } else {
        filled.set(28);
      }

      byteB0.add(b);

      return this;
    }

    TraceBuilder byteB1(final UnsignedByte b) {
      if (filled.get(27)) {
        throw new IllegalStateException("BYTE_B_1 already set");
      } else {
        filled.set(27);
      }

      byteB1.add(b);

      return this;
    }

    TraceBuilder byteB2(final UnsignedByte b) {
      if (filled.get(87)) {
        throw new IllegalStateException("BYTE_B_2 already set");
      } else {
        filled.set(87);
      }

      byteB2.add(b);

      return this;
    }

    TraceBuilder byteB3(final UnsignedByte b) {
      if (filled.get(59)) {
        throw new IllegalStateException("BYTE_B_3 already set");
      } else {
        filled.set(59);
      }

      byteB3.add(b);

      return this;
    }

    TraceBuilder byteC0(final UnsignedByte b) {
      if (filled.get(45)) {
        throw new IllegalStateException("BYTE_C_0 already set");
      } else {
        filled.set(45);
      }

      byteC0.add(b);

      return this;
    }

    TraceBuilder byteC1(final UnsignedByte b) {
      if (filled.get(65)) {
        throw new IllegalStateException("BYTE_C_1 already set");
      } else {
        filled.set(65);
      }

      byteC1.add(b);

      return this;
    }

    TraceBuilder byteC2(final UnsignedByte b) {
      if (filled.get(111)) {
        throw new IllegalStateException("BYTE_C_2 already set");
      } else {
        filled.set(111);
      }

      byteC2.add(b);

      return this;
    }

    TraceBuilder byteC3(final UnsignedByte b) {
      if (filled.get(82)) {
        throw new IllegalStateException("BYTE_C_3 already set");
      } else {
        filled.set(82);
      }

      byteC3.add(b);

      return this;
    }

    TraceBuilder byteDelta0(final UnsignedByte b) {
      if (filled.get(92)) {
        throw new IllegalStateException("BYTE_DELTA_0 already set");
      } else {
        filled.set(92);
      }

      byteDelta0.add(b);

      return this;
    }

    TraceBuilder byteDelta1(final UnsignedByte b) {
      if (filled.get(18)) {
        throw new IllegalStateException("BYTE_DELTA_1 already set");
      } else {
        filled.set(18);
      }

      byteDelta1.add(b);

      return this;
    }

    TraceBuilder byteDelta2(final UnsignedByte b) {
      if (filled.get(74)) {
        throw new IllegalStateException("BYTE_DELTA_2 already set");
      } else {
        filled.set(74);
      }

      byteDelta2.add(b);

      return this;
    }

    TraceBuilder byteDelta3(final UnsignedByte b) {
      if (filled.get(37)) {
        throw new IllegalStateException("BYTE_DELTA_3 already set");
      } else {
        filled.set(37);
      }

      byteDelta3.add(b);

      return this;
    }

    TraceBuilder byteH0(final UnsignedByte b) {
      if (filled.get(42)) {
        throw new IllegalStateException("BYTE_H_0 already set");
      } else {
        filled.set(42);
      }

      byteH0.add(b);

      return this;
    }

    TraceBuilder byteH1(final UnsignedByte b) {
      if (filled.get(12)) {
        throw new IllegalStateException("BYTE_H_1 already set");
      } else {
        filled.set(12);
      }

      byteH1.add(b);

      return this;
    }

    TraceBuilder byteH2(final UnsignedByte b) {
      if (filled.get(88)) {
        throw new IllegalStateException("BYTE_H_2 already set");
      } else {
        filled.set(88);
      }

      byteH2.add(b);

      return this;
    }

    TraceBuilder byteH3(final UnsignedByte b) {
      if (filled.get(75)) {
        throw new IllegalStateException("BYTE_H_3 already set");
      } else {
        filled.set(75);
      }

      byteH3.add(b);

      return this;
    }

    TraceBuilder byteH4(final UnsignedByte b) {
      if (filled.get(55)) {
        throw new IllegalStateException("BYTE_H_4 already set");
      } else {
        filled.set(55);
      }

      byteH4.add(b);

      return this;
    }

    TraceBuilder byteH5(final UnsignedByte b) {
      if (filled.get(99)) {
        throw new IllegalStateException("BYTE_H_5 already set");
      } else {
        filled.set(99);
      }

      byteH5.add(b);

      return this;
    }

    TraceBuilder byteI0(final UnsignedByte b) {
      if (filled.get(80)) {
        throw new IllegalStateException("BYTE_I_0 already set");
      } else {
        filled.set(80);
      }

      byteI0.add(b);

      return this;
    }

    TraceBuilder byteI1(final UnsignedByte b) {
      if (filled.get(31)) {
        throw new IllegalStateException("BYTE_I_1 already set");
      } else {
        filled.set(31);
      }

      byteI1.add(b);

      return this;
    }

    TraceBuilder byteI2(final UnsignedByte b) {
      if (filled.get(78)) {
        throw new IllegalStateException("BYTE_I_2 already set");
      } else {
        filled.set(78);
      }

      byteI2.add(b);

      return this;
    }

    TraceBuilder byteI3(final UnsignedByte b) {
      if (filled.get(19)) {
        throw new IllegalStateException("BYTE_I_3 already set");
      } else {
        filled.set(19);
      }

      byteI3.add(b);

      return this;
    }

    TraceBuilder byteI4(final UnsignedByte b) {
      if (filled.get(72)) {
        throw new IllegalStateException("BYTE_I_4 already set");
      } else {
        filled.set(72);
      }

      byteI4.add(b);

      return this;
    }

    TraceBuilder byteI5(final UnsignedByte b) {
      if (filled.get(52)) {
        throw new IllegalStateException("BYTE_I_5 already set");
      } else {
        filled.set(52);
      }

      byteI5.add(b);

      return this;
    }

    TraceBuilder byteI6(final UnsignedByte b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BYTE_I_6 already set");
      } else {
        filled.set(10);
      }

      byteI6.add(b);

      return this;
    }

    TraceBuilder byteJ0(final UnsignedByte b) {
      if (filled.get(84)) {
        throw new IllegalStateException("BYTE_J_0 already set");
      } else {
        filled.set(84);
      }

      byteJ0.add(b);

      return this;
    }

    TraceBuilder byteJ1(final UnsignedByte b) {
      if (filled.get(43)) {
        throw new IllegalStateException("BYTE_J_1 already set");
      } else {
        filled.set(43);
      }

      byteJ1.add(b);

      return this;
    }

    TraceBuilder byteJ2(final UnsignedByte b) {
      if (filled.get(21)) {
        throw new IllegalStateException("BYTE_J_2 already set");
      } else {
        filled.set(21);
      }

      byteJ2.add(b);

      return this;
    }

    TraceBuilder byteJ3(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BYTE_J_3 already set");
      } else {
        filled.set(23);
      }

      byteJ3.add(b);

      return this;
    }

    TraceBuilder byteJ4(final UnsignedByte b) {
      if (filled.get(60)) {
        throw new IllegalStateException("BYTE_J_4 already set");
      } else {
        filled.set(60);
      }

      byteJ4.add(b);

      return this;
    }

    TraceBuilder byteJ5(final UnsignedByte b) {
      if (filled.get(61)) {
        throw new IllegalStateException("BYTE_J_5 already set");
      } else {
        filled.set(61);
      }

      byteJ5.add(b);

      return this;
    }

    TraceBuilder byteJ6(final UnsignedByte b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BYTE_J_6 already set");
      } else {
        filled.set(20);
      }

      byteJ6.add(b);

      return this;
    }

    TraceBuilder byteJ7(final UnsignedByte b) {
      if (filled.get(106)) {
        throw new IllegalStateException("BYTE_J_7 already set");
      } else {
        filled.set(106);
      }

      byteJ7.add(b);

      return this;
    }

    TraceBuilder byteQ0(final UnsignedByte b) {
      if (filled.get(41)) {
        throw new IllegalStateException("BYTE_Q_0 already set");
      } else {
        filled.set(41);
      }

      byteQ0.add(b);

      return this;
    }

    TraceBuilder byteQ1(final UnsignedByte b) {
      if (filled.get(5)) {
        throw new IllegalStateException("BYTE_Q_1 already set");
      } else {
        filled.set(5);
      }

      byteQ1.add(b);

      return this;
    }

    TraceBuilder byteQ2(final UnsignedByte b) {
      if (filled.get(22)) {
        throw new IllegalStateException("BYTE_Q_2 already set");
      } else {
        filled.set(22);
      }

      byteQ2.add(b);

      return this;
    }

    TraceBuilder byteQ3(final UnsignedByte b) {
      if (filled.get(58)) {
        throw new IllegalStateException("BYTE_Q_3 already set");
      } else {
        filled.set(58);
      }

      byteQ3.add(b);

      return this;
    }

    TraceBuilder byteQ4(final UnsignedByte b) {
      if (filled.get(107)) {
        throw new IllegalStateException("BYTE_Q_4 already set");
      } else {
        filled.set(107);
      }

      byteQ4.add(b);

      return this;
    }

    TraceBuilder byteQ5(final UnsignedByte b) {
      if (filled.get(7)) {
        throw new IllegalStateException("BYTE_Q_5 already set");
      } else {
        filled.set(7);
      }

      byteQ5.add(b);

      return this;
    }

    TraceBuilder byteQ6(final UnsignedByte b) {
      if (filled.get(117)) {
        throw new IllegalStateException("BYTE_Q_6 already set");
      } else {
        filled.set(117);
      }

      byteQ6.add(b);

      return this;
    }

    TraceBuilder byteQ7(final UnsignedByte b) {
      if (filled.get(112)) {
        throw new IllegalStateException("BYTE_Q_7 already set");
      } else {
        filled.set(112);
      }

      byteQ7.add(b);

      return this;
    }

    TraceBuilder byteR0(final UnsignedByte b) {
      if (filled.get(35)) {
        throw new IllegalStateException("BYTE_R_0 already set");
      } else {
        filled.set(35);
      }

      byteR0.add(b);

      return this;
    }

    TraceBuilder byteR1(final UnsignedByte b) {
      if (filled.get(2)) {
        throw new IllegalStateException("BYTE_R_1 already set");
      } else {
        filled.set(2);
      }

      byteR1.add(b);

      return this;
    }

    TraceBuilder byteR2(final UnsignedByte b) {
      if (filled.get(47)) {
        throw new IllegalStateException("BYTE_R_2 already set");
      } else {
        filled.set(47);
      }

      byteR2.add(b);

      return this;
    }

    TraceBuilder byteR3(final UnsignedByte b) {
      if (filled.get(9)) {
        throw new IllegalStateException("BYTE_R_3 already set");
      } else {
        filled.set(9);
      }

      byteR3.add(b);

      return this;
    }

    TraceBuilder cmp(final Boolean b) {
      if (filled.get(103)) {
        throw new IllegalStateException("CMP already set");
      } else {
        filled.set(103);
      }

      cmp.add(b);

      return this;
    }

    TraceBuilder ct(final BigInteger b) {
      if (filled.get(110)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(110);
      }

      ct.add(b);

      return this;
    }

    TraceBuilder inst(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(108);
      }

      inst.add(b);

      return this;
    }

    TraceBuilder ofH(final Boolean b) {
      if (filled.get(39)) {
        throw new IllegalStateException("OF_H already set");
      } else {
        filled.set(39);
      }

      ofH.add(b);

      return this;
    }

    TraceBuilder ofI(final Boolean b) {
      if (filled.get(100)) {
        throw new IllegalStateException("OF_I already set");
      } else {
        filled.set(100);
      }

      ofI.add(b);

      return this;
    }

    TraceBuilder ofJ(final Boolean b) {
      if (filled.get(96)) {
        throw new IllegalStateException("OF_J already set");
      } else {
        filled.set(96);
      }

      ofJ.add(b);

      return this;
    }

    TraceBuilder ofRes(final Boolean b) {
      if (filled.get(63)) {
        throw new IllegalStateException("OF_RES already set");
      } else {
        filled.set(63);
      }

      ofRes.add(b);

      return this;
    }

    TraceBuilder oli(final Boolean b) {
      if (filled.get(0)) {
        throw new IllegalStateException("OLI already set");
      } else {
        filled.set(0);
      }

      oli.add(b);

      return this;
    }

    TraceBuilder resHi(final BigInteger b) {
      if (filled.get(71)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(71);
      }

      resHi.add(b);

      return this;
    }

    TraceBuilder resLo(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(109);
      }

      resLo.add(b);

      return this;
    }

    TraceBuilder stamp(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(13);
      }

      stamp.add(b);

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

    TraceBuilder setAccH3At(final BigInteger b, int i) {
      accH3.set(i, b);

      return this;
    }

    TraceBuilder setAccH4At(final BigInteger b, int i) {
      accH4.set(i, b);

      return this;
    }

    TraceBuilder setAccH5At(final BigInteger b, int i) {
      accH5.set(i, b);

      return this;
    }

    TraceBuilder setAccI0At(final BigInteger b, int i) {
      accI0.set(i, b);

      return this;
    }

    TraceBuilder setAccI1At(final BigInteger b, int i) {
      accI1.set(i, b);

      return this;
    }

    TraceBuilder setAccI2At(final BigInteger b, int i) {
      accI2.set(i, b);

      return this;
    }

    TraceBuilder setAccI3At(final BigInteger b, int i) {
      accI3.set(i, b);

      return this;
    }

    TraceBuilder setAccI4At(final BigInteger b, int i) {
      accI4.set(i, b);

      return this;
    }

    TraceBuilder setAccI5At(final BigInteger b, int i) {
      accI5.set(i, b);

      return this;
    }

    TraceBuilder setAccI6At(final BigInteger b, int i) {
      accI6.set(i, b);

      return this;
    }

    TraceBuilder setAccJ0At(final BigInteger b, int i) {
      accJ0.set(i, b);

      return this;
    }

    TraceBuilder setAccJ1At(final BigInteger b, int i) {
      accJ1.set(i, b);

      return this;
    }

    TraceBuilder setAccJ2At(final BigInteger b, int i) {
      accJ2.set(i, b);

      return this;
    }

    TraceBuilder setAccJ3At(final BigInteger b, int i) {
      accJ3.set(i, b);

      return this;
    }

    TraceBuilder setAccJ4At(final BigInteger b, int i) {
      accJ4.set(i, b);

      return this;
    }

    TraceBuilder setAccJ5At(final BigInteger b, int i) {
      accJ5.set(i, b);

      return this;
    }

    TraceBuilder setAccJ6At(final BigInteger b, int i) {
      accJ6.set(i, b);

      return this;
    }

    TraceBuilder setAccJ7At(final BigInteger b, int i) {
      accJ7.set(i, b);

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

    TraceBuilder setAccQ4At(final BigInteger b, int i) {
      accQ4.set(i, b);

      return this;
    }

    TraceBuilder setAccQ5At(final BigInteger b, int i) {
      accQ5.set(i, b);

      return this;
    }

    TraceBuilder setAccQ6At(final BigInteger b, int i) {
      accQ6.set(i, b);

      return this;
    }

    TraceBuilder setAccQ7At(final BigInteger b, int i) {
      accQ7.set(i, b);

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

    TraceBuilder setArg3HiAt(final BigInteger b, int i) {
      arg3Hi.set(i, b);

      return this;
    }

    TraceBuilder setArg3LoAt(final BigInteger b, int i) {
      arg3Lo.set(i, b);

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

    TraceBuilder setByteH3At(final UnsignedByte b, int i) {
      byteH3.set(i, b);

      return this;
    }

    TraceBuilder setByteH4At(final UnsignedByte b, int i) {
      byteH4.set(i, b);

      return this;
    }

    TraceBuilder setByteH5At(final UnsignedByte b, int i) {
      byteH5.set(i, b);

      return this;
    }

    TraceBuilder setByteI0At(final UnsignedByte b, int i) {
      byteI0.set(i, b);

      return this;
    }

    TraceBuilder setByteI1At(final UnsignedByte b, int i) {
      byteI1.set(i, b);

      return this;
    }

    TraceBuilder setByteI2At(final UnsignedByte b, int i) {
      byteI2.set(i, b);

      return this;
    }

    TraceBuilder setByteI3At(final UnsignedByte b, int i) {
      byteI3.set(i, b);

      return this;
    }

    TraceBuilder setByteI4At(final UnsignedByte b, int i) {
      byteI4.set(i, b);

      return this;
    }

    TraceBuilder setByteI5At(final UnsignedByte b, int i) {
      byteI5.set(i, b);

      return this;
    }

    TraceBuilder setByteI6At(final UnsignedByte b, int i) {
      byteI6.set(i, b);

      return this;
    }

    TraceBuilder setByteJ0At(final UnsignedByte b, int i) {
      byteJ0.set(i, b);

      return this;
    }

    TraceBuilder setByteJ1At(final UnsignedByte b, int i) {
      byteJ1.set(i, b);

      return this;
    }

    TraceBuilder setByteJ2At(final UnsignedByte b, int i) {
      byteJ2.set(i, b);

      return this;
    }

    TraceBuilder setByteJ3At(final UnsignedByte b, int i) {
      byteJ3.set(i, b);

      return this;
    }

    TraceBuilder setByteJ4At(final UnsignedByte b, int i) {
      byteJ4.set(i, b);

      return this;
    }

    TraceBuilder setByteJ5At(final UnsignedByte b, int i) {
      byteJ5.set(i, b);

      return this;
    }

    TraceBuilder setByteJ6At(final UnsignedByte b, int i) {
      byteJ6.set(i, b);

      return this;
    }

    TraceBuilder setByteJ7At(final UnsignedByte b, int i) {
      byteJ7.set(i, b);

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

    TraceBuilder setByteQ4At(final UnsignedByte b, int i) {
      byteQ4.set(i, b);

      return this;
    }

    TraceBuilder setByteQ5At(final UnsignedByte b, int i) {
      byteQ5.set(i, b);

      return this;
    }

    TraceBuilder setByteQ6At(final UnsignedByte b, int i) {
      byteQ6.set(i, b);

      return this;
    }

    TraceBuilder setByteQ7At(final UnsignedByte b, int i) {
      byteQ7.set(i, b);

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

    TraceBuilder setCmpAt(final Boolean b, int i) {
      cmp.set(i, b);

      return this;
    }

    TraceBuilder setCtAt(final BigInteger b, int i) {
      ct.set(i, b);

      return this;
    }

    TraceBuilder setInstAt(final BigInteger b, int i) {
      inst.set(i, b);

      return this;
    }

    TraceBuilder setOfHAt(final Boolean b, int i) {
      ofH.set(i, b);

      return this;
    }

    TraceBuilder setOfIAt(final Boolean b, int i) {
      ofI.set(i, b);

      return this;
    }

    TraceBuilder setOfJAt(final Boolean b, int i) {
      ofJ.set(i, b);

      return this;
    }

    TraceBuilder setOfResAt(final Boolean b, int i) {
      ofRes.set(i, b);

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

    TraceBuilder setAccH3Relative(final BigInteger b, int i) {
      accH3.set(accH3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccH4Relative(final BigInteger b, int i) {
      accH4.set(accH4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccH5Relative(final BigInteger b, int i) {
      accH5.set(accH5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI0Relative(final BigInteger b, int i) {
      accI0.set(accI0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI1Relative(final BigInteger b, int i) {
      accI1.set(accI1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI2Relative(final BigInteger b, int i) {
      accI2.set(accI2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI3Relative(final BigInteger b, int i) {
      accI3.set(accI3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI4Relative(final BigInteger b, int i) {
      accI4.set(accI4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI5Relative(final BigInteger b, int i) {
      accI5.set(accI5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccI6Relative(final BigInteger b, int i) {
      accI6.set(accI6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ0Relative(final BigInteger b, int i) {
      accJ0.set(accJ0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ1Relative(final BigInteger b, int i) {
      accJ1.set(accJ1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ2Relative(final BigInteger b, int i) {
      accJ2.set(accJ2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ3Relative(final BigInteger b, int i) {
      accJ3.set(accJ3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ4Relative(final BigInteger b, int i) {
      accJ4.set(accJ4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ5Relative(final BigInteger b, int i) {
      accJ5.set(accJ5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ6Relative(final BigInteger b, int i) {
      accJ6.set(accJ6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccJ7Relative(final BigInteger b, int i) {
      accJ7.set(accJ7.size() - 1 - i, b);

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

    TraceBuilder setAccQ4Relative(final BigInteger b, int i) {
      accQ4.set(accQ4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccQ5Relative(final BigInteger b, int i) {
      accQ5.set(accQ5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccQ6Relative(final BigInteger b, int i) {
      accQ6.set(accQ6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccQ7Relative(final BigInteger b, int i) {
      accQ7.set(accQ7.size() - 1 - i, b);

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

    TraceBuilder setArg3HiRelative(final BigInteger b, int i) {
      arg3Hi.set(arg3Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg3LoRelative(final BigInteger b, int i) {
      arg3Lo.set(arg3Lo.size() - 1 - i, b);

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

    TraceBuilder setByteH3Relative(final UnsignedByte b, int i) {
      byteH3.set(byteH3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteH4Relative(final UnsignedByte b, int i) {
      byteH4.set(byteH4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteH5Relative(final UnsignedByte b, int i) {
      byteH5.set(byteH5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI0Relative(final UnsignedByte b, int i) {
      byteI0.set(byteI0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI1Relative(final UnsignedByte b, int i) {
      byteI1.set(byteI1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI2Relative(final UnsignedByte b, int i) {
      byteI2.set(byteI2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI3Relative(final UnsignedByte b, int i) {
      byteI3.set(byteI3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI4Relative(final UnsignedByte b, int i) {
      byteI4.set(byteI4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI5Relative(final UnsignedByte b, int i) {
      byteI5.set(byteI5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteI6Relative(final UnsignedByte b, int i) {
      byteI6.set(byteI6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ0Relative(final UnsignedByte b, int i) {
      byteJ0.set(byteJ0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ1Relative(final UnsignedByte b, int i) {
      byteJ1.set(byteJ1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ2Relative(final UnsignedByte b, int i) {
      byteJ2.set(byteJ2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ3Relative(final UnsignedByte b, int i) {
      byteJ3.set(byteJ3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ4Relative(final UnsignedByte b, int i) {
      byteJ4.set(byteJ4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ5Relative(final UnsignedByte b, int i) {
      byteJ5.set(byteJ5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ6Relative(final UnsignedByte b, int i) {
      byteJ6.set(byteJ6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteJ7Relative(final UnsignedByte b, int i) {
      byteJ7.set(byteJ7.size() - 1 - i, b);

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

    TraceBuilder setByteQ4Relative(final UnsignedByte b, int i) {
      byteQ4.set(byteQ4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteQ5Relative(final UnsignedByte b, int i) {
      byteQ5.set(byteQ5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteQ6Relative(final UnsignedByte b, int i) {
      byteQ6.set(byteQ6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByteQ7Relative(final UnsignedByte b, int i) {
      byteQ7.set(byteQ7.size() - 1 - i, b);

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

    TraceBuilder setCmpRelative(final Boolean b, int i) {
      cmp.set(cmp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCtRelative(final BigInteger b, int i) {
      ct.set(ct.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInstRelative(final BigInteger b, int i) {
      inst.set(inst.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOfHRelative(final Boolean b, int i) {
      ofH.set(ofH.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOfIRelative(final Boolean b, int i) {
      ofI.set(ofI.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOfJRelative(final Boolean b, int i) {
      ofJ.set(ofJ.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOfResRelative(final Boolean b, int i) {
      ofRes.set(ofRes.size() - 1 - i, b);

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
      if (!filled.get(66)) {
        throw new IllegalStateException("ACC_A_0 has not been filled");
      }

      if (!filled.get(86)) {
        throw new IllegalStateException("ACC_A_1 has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("ACC_A_2 has not been filled");
      }

      if (!filled.get(79)) {
        throw new IllegalStateException("ACC_A_3 has not been filled");
      }

      if (!filled.get(70)) {
        throw new IllegalStateException("ACC_B_0 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("ACC_B_1 has not been filled");
      }

      if (!filled.get(98)) {
        throw new IllegalStateException("ACC_B_2 has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("ACC_B_3 has not been filled");
      }

      if (!filled.get(83)) {
        throw new IllegalStateException("ACC_C_0 has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("ACC_C_1 has not been filled");
      }

      if (!filled.get(62)) {
        throw new IllegalStateException("ACC_C_2 has not been filled");
      }

      if (!filled.get(114)) {
        throw new IllegalStateException("ACC_C_3 has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("ACC_DELTA_0 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("ACC_DELTA_1 has not been filled");
      }

      if (!filled.get(113)) {
        throw new IllegalStateException("ACC_DELTA_2 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("ACC_DELTA_3 has not been filled");
      }

      if (!filled.get(95)) {
        throw new IllegalStateException("ACC_H_0 has not been filled");
      }

      if (!filled.get(67)) {
        throw new IllegalStateException("ACC_H_1 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("ACC_H_2 has not been filled");
      }

      if (!filled.get(76)) {
        throw new IllegalStateException("ACC_H_3 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("ACC_H_4 has not been filled");
      }

      if (!filled.get(81)) {
        throw new IllegalStateException("ACC_H_5 has not been filled");
      }

      if (!filled.get(68)) {
        throw new IllegalStateException("ACC_I_0 has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("ACC_I_1 has not been filled");
      }

      if (!filled.get(105)) {
        throw new IllegalStateException("ACC_I_2 has not been filled");
      }

      if (!filled.get(77)) {
        throw new IllegalStateException("ACC_I_3 has not been filled");
      }

      if (!filled.get(73)) {
        throw new IllegalStateException("ACC_I_4 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("ACC_I_5 has not been filled");
      }

      if (!filled.get(69)) {
        throw new IllegalStateException("ACC_I_6 has not been filled");
      }

      if (!filled.get(93)) {
        throw new IllegalStateException("ACC_J_0 has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ACC_J_1 has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("ACC_J_2 has not been filled");
      }

      if (!filled.get(64)) {
        throw new IllegalStateException("ACC_J_3 has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("ACC_J_4 has not been filled");
      }

      if (!filled.get(85)) {
        throw new IllegalStateException("ACC_J_5 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("ACC_J_6 has not been filled");
      }

      if (!filled.get(89)) {
        throw new IllegalStateException("ACC_J_7 has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("ACC_Q_0 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_Q_1 has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("ACC_Q_2 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_Q_3 has not been filled");
      }

      if (!filled.get(94)) {
        throw new IllegalStateException("ACC_Q_4 has not been filled");
      }

      if (!filled.get(116)) {
        throw new IllegalStateException("ACC_Q_5 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("ACC_Q_6 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("ACC_Q_7 has not been filled");
      }

      if (!filled.get(104)) {
        throw new IllegalStateException("ACC_R_0 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("ACC_R_1 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("ACC_R_2 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("ACC_R_3 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(115)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("ARG_3_HI has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ARG_3_LO has not been filled");
      }

      if (!filled.get(101)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(97)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(91)) {
        throw new IllegalStateException("BYTE_A_0 has not been filled");
      }

      if (!filled.get(90)) {
        throw new IllegalStateException("BYTE_A_1 has not been filled");
      }

      if (!filled.get(102)) {
        throw new IllegalStateException("BYTE_A_2 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("BYTE_A_3 has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("BYTE_B_0 has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("BYTE_B_1 has not been filled");
      }

      if (!filled.get(87)) {
        throw new IllegalStateException("BYTE_B_2 has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("BYTE_B_3 has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("BYTE_C_0 has not been filled");
      }

      if (!filled.get(65)) {
        throw new IllegalStateException("BYTE_C_1 has not been filled");
      }

      if (!filled.get(111)) {
        throw new IllegalStateException("BYTE_C_2 has not been filled");
      }

      if (!filled.get(82)) {
        throw new IllegalStateException("BYTE_C_3 has not been filled");
      }

      if (!filled.get(92)) {
        throw new IllegalStateException("BYTE_DELTA_0 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("BYTE_DELTA_1 has not been filled");
      }

      if (!filled.get(74)) {
        throw new IllegalStateException("BYTE_DELTA_2 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("BYTE_DELTA_3 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("BYTE_H_0 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("BYTE_H_1 has not been filled");
      }

      if (!filled.get(88)) {
        throw new IllegalStateException("BYTE_H_2 has not been filled");
      }

      if (!filled.get(75)) {
        throw new IllegalStateException("BYTE_H_3 has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("BYTE_H_4 has not been filled");
      }

      if (!filled.get(99)) {
        throw new IllegalStateException("BYTE_H_5 has not been filled");
      }

      if (!filled.get(80)) {
        throw new IllegalStateException("BYTE_I_0 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("BYTE_I_1 has not been filled");
      }

      if (!filled.get(78)) {
        throw new IllegalStateException("BYTE_I_2 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("BYTE_I_3 has not been filled");
      }

      if (!filled.get(72)) {
        throw new IllegalStateException("BYTE_I_4 has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("BYTE_I_5 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BYTE_I_6 has not been filled");
      }

      if (!filled.get(84)) {
        throw new IllegalStateException("BYTE_J_0 has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("BYTE_J_1 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("BYTE_J_2 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BYTE_J_3 has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("BYTE_J_4 has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException("BYTE_J_5 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BYTE_J_6 has not been filled");
      }

      if (!filled.get(106)) {
        throw new IllegalStateException("BYTE_J_7 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("BYTE_Q_0 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("BYTE_Q_1 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("BYTE_Q_2 has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("BYTE_Q_3 has not been filled");
      }

      if (!filled.get(107)) {
        throw new IllegalStateException("BYTE_Q_4 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("BYTE_Q_5 has not been filled");
      }

      if (!filled.get(117)) {
        throw new IllegalStateException("BYTE_Q_6 has not been filled");
      }

      if (!filled.get(112)) {
        throw new IllegalStateException("BYTE_Q_7 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("BYTE_R_0 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("BYTE_R_1 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("BYTE_R_2 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("BYTE_R_3 has not been filled");
      }

      if (!filled.get(103)) {
        throw new IllegalStateException("CMP has not been filled");
      }

      if (!filled.get(110)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(108)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("OF_H has not been filled");
      }

      if (!filled.get(100)) {
        throw new IllegalStateException("OF_I has not been filled");
      }

      if (!filled.get(96)) {
        throw new IllegalStateException("OF_J has not been filled");
      }

      if (!filled.get(63)) {
        throw new IllegalStateException("OF_RES has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("OLI has not been filled");
      }

      if (!filled.get(71)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(109)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(13)) {
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
          accDelta0,
          accDelta1,
          accDelta2,
          accDelta3,
          accH0,
          accH1,
          accH2,
          accH3,
          accH4,
          accH5,
          accI0,
          accI1,
          accI2,
          accI3,
          accI4,
          accI5,
          accI6,
          accJ0,
          accJ1,
          accJ2,
          accJ3,
          accJ4,
          accJ5,
          accJ6,
          accJ7,
          accQ0,
          accQ1,
          accQ2,
          accQ3,
          accQ4,
          accQ5,
          accQ6,
          accQ7,
          accR0,
          accR1,
          accR2,
          accR3,
          arg1Hi,
          arg1Lo,
          arg2Hi,
          arg2Lo,
          arg3Hi,
          arg3Lo,
          bit1,
          bit2,
          bit3,
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
          byteDelta0,
          byteDelta1,
          byteDelta2,
          byteDelta3,
          byteH0,
          byteH1,
          byteH2,
          byteH3,
          byteH4,
          byteH5,
          byteI0,
          byteI1,
          byteI2,
          byteI3,
          byteI4,
          byteI5,
          byteI6,
          byteJ0,
          byteJ1,
          byteJ2,
          byteJ3,
          byteJ4,
          byteJ5,
          byteJ6,
          byteJ7,
          byteQ0,
          byteQ1,
          byteQ2,
          byteQ3,
          byteQ4,
          byteQ5,
          byteQ6,
          byteQ7,
          byteR0,
          byteR1,
          byteR2,
          byteR3,
          cmp,
          ct,
          inst,
          ofH,
          ofI,
          ofJ,
          ofRes,
          oli,
          resHi,
          resLo,
          stamp);
    }
  }
}
