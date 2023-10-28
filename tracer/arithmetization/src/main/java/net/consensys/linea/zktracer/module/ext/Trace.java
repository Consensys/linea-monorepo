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

    @JsonProperty("ACC_DELTA_0")
    private final List<BigInteger> accDelta0;

    @JsonProperty("ACC_DELTA_1")
    private final List<BigInteger> accDelta1;

    @JsonProperty("ACC_DELTA_2")
    private final List<BigInteger> accDelta2;

    @JsonProperty("ACC_DELTA_3")
    private final List<BigInteger> accDelta3;

    @JsonProperty("ACC_H_0")
    private final List<BigInteger> accH0;

    @JsonProperty("ACC_H_1")
    private final List<BigInteger> accH1;

    @JsonProperty("ACC_H_2")
    private final List<BigInteger> accH2;

    @JsonProperty("ACC_H_3")
    private final List<BigInteger> accH3;

    @JsonProperty("ACC_H_4")
    private final List<BigInteger> accH4;

    @JsonProperty("ACC_H_5")
    private final List<BigInteger> accH5;

    @JsonProperty("ACC_I_0")
    private final List<BigInteger> accI0;

    @JsonProperty("ACC_I_1")
    private final List<BigInteger> accI1;

    @JsonProperty("ACC_I_2")
    private final List<BigInteger> accI2;

    @JsonProperty("ACC_I_3")
    private final List<BigInteger> accI3;

    @JsonProperty("ACC_I_4")
    private final List<BigInteger> accI4;

    @JsonProperty("ACC_I_5")
    private final List<BigInteger> accI5;

    @JsonProperty("ACC_I_6")
    private final List<BigInteger> accI6;

    @JsonProperty("ACC_J_0")
    private final List<BigInteger> accJ0;

    @JsonProperty("ACC_J_1")
    private final List<BigInteger> accJ1;

    @JsonProperty("ACC_J_2")
    private final List<BigInteger> accJ2;

    @JsonProperty("ACC_J_3")
    private final List<BigInteger> accJ3;

    @JsonProperty("ACC_J_4")
    private final List<BigInteger> accJ4;

    @JsonProperty("ACC_J_5")
    private final List<BigInteger> accJ5;

    @JsonProperty("ACC_J_6")
    private final List<BigInteger> accJ6;

    @JsonProperty("ACC_J_7")
    private final List<BigInteger> accJ7;

    @JsonProperty("ACC_Q_0")
    private final List<BigInteger> accQ0;

    @JsonProperty("ACC_Q_1")
    private final List<BigInteger> accQ1;

    @JsonProperty("ACC_Q_2")
    private final List<BigInteger> accQ2;

    @JsonProperty("ACC_Q_3")
    private final List<BigInteger> accQ3;

    @JsonProperty("ACC_Q_4")
    private final List<BigInteger> accQ4;

    @JsonProperty("ACC_Q_5")
    private final List<BigInteger> accQ5;

    @JsonProperty("ACC_Q_6")
    private final List<BigInteger> accQ6;

    @JsonProperty("ACC_Q_7")
    private final List<BigInteger> accQ7;

    @JsonProperty("ACC_R_0")
    private final List<BigInteger> accR0;

    @JsonProperty("ACC_R_1")
    private final List<BigInteger> accR1;

    @JsonProperty("ACC_R_2")
    private final List<BigInteger> accR2;

    @JsonProperty("ACC_R_3")
    private final List<BigInteger> accR3;

    @JsonProperty("ARG_1_HI")
    private final List<BigInteger> arg1Hi;

    @JsonProperty("ARG_1_LO")
    private final List<BigInteger> arg1Lo;

    @JsonProperty("ARG_2_HI")
    private final List<BigInteger> arg2Hi;

    @JsonProperty("ARG_2_LO")
    private final List<BigInteger> arg2Lo;

    @JsonProperty("ARG_3_HI")
    private final List<BigInteger> arg3Hi;

    @JsonProperty("ARG_3_LO")
    private final List<BigInteger> arg3Lo;

    @JsonProperty("BIT_1")
    private final List<Boolean> bit1;

    @JsonProperty("BIT_2")
    private final List<Boolean> bit2;

    @JsonProperty("BIT_3")
    private final List<Boolean> bit3;

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

    @JsonProperty("BYTE_DELTA_0")
    private final List<UnsignedByte> byteDelta0;

    @JsonProperty("BYTE_DELTA_1")
    private final List<UnsignedByte> byteDelta1;

    @JsonProperty("BYTE_DELTA_2")
    private final List<UnsignedByte> byteDelta2;

    @JsonProperty("BYTE_DELTA_3")
    private final List<UnsignedByte> byteDelta3;

    @JsonProperty("BYTE_H_0")
    private final List<UnsignedByte> byteH0;

    @JsonProperty("BYTE_H_1")
    private final List<UnsignedByte> byteH1;

    @JsonProperty("BYTE_H_2")
    private final List<UnsignedByte> byteH2;

    @JsonProperty("BYTE_H_3")
    private final List<UnsignedByte> byteH3;

    @JsonProperty("BYTE_H_4")
    private final List<UnsignedByte> byteH4;

    @JsonProperty("BYTE_H_5")
    private final List<UnsignedByte> byteH5;

    @JsonProperty("BYTE_I_0")
    private final List<UnsignedByte> byteI0;

    @JsonProperty("BYTE_I_1")
    private final List<UnsignedByte> byteI1;

    @JsonProperty("BYTE_I_2")
    private final List<UnsignedByte> byteI2;

    @JsonProperty("BYTE_I_3")
    private final List<UnsignedByte> byteI3;

    @JsonProperty("BYTE_I_4")
    private final List<UnsignedByte> byteI4;

    @JsonProperty("BYTE_I_5")
    private final List<UnsignedByte> byteI5;

    @JsonProperty("BYTE_I_6")
    private final List<UnsignedByte> byteI6;

    @JsonProperty("BYTE_J_0")
    private final List<UnsignedByte> byteJ0;

    @JsonProperty("BYTE_J_1")
    private final List<UnsignedByte> byteJ1;

    @JsonProperty("BYTE_J_2")
    private final List<UnsignedByte> byteJ2;

    @JsonProperty("BYTE_J_3")
    private final List<UnsignedByte> byteJ3;

    @JsonProperty("BYTE_J_4")
    private final List<UnsignedByte> byteJ4;

    @JsonProperty("BYTE_J_5")
    private final List<UnsignedByte> byteJ5;

    @JsonProperty("BYTE_J_6")
    private final List<UnsignedByte> byteJ6;

    @JsonProperty("BYTE_J_7")
    private final List<UnsignedByte> byteJ7;

    @JsonProperty("BYTE_Q_0")
    private final List<UnsignedByte> byteQ0;

    @JsonProperty("BYTE_Q_1")
    private final List<UnsignedByte> byteQ1;

    @JsonProperty("BYTE_Q_2")
    private final List<UnsignedByte> byteQ2;

    @JsonProperty("BYTE_Q_3")
    private final List<UnsignedByte> byteQ3;

    @JsonProperty("BYTE_Q_4")
    private final List<UnsignedByte> byteQ4;

    @JsonProperty("BYTE_Q_5")
    private final List<UnsignedByte> byteQ5;

    @JsonProperty("BYTE_Q_6")
    private final List<UnsignedByte> byteQ6;

    @JsonProperty("BYTE_Q_7")
    private final List<UnsignedByte> byteQ7;

    @JsonProperty("BYTE_R_0")
    private final List<UnsignedByte> byteR0;

    @JsonProperty("BYTE_R_1")
    private final List<UnsignedByte> byteR1;

    @JsonProperty("BYTE_R_2")
    private final List<UnsignedByte> byteR2;

    @JsonProperty("BYTE_R_3")
    private final List<UnsignedByte> byteR3;

    @JsonProperty("CMP")
    private final List<Boolean> cmp;

    @JsonProperty("CT")
    private final List<BigInteger> ct;

    @JsonProperty("INST")
    private final List<BigInteger> inst;

    @JsonProperty("OF_H")
    private final List<Boolean> ofH;

    @JsonProperty("OF_I")
    private final List<Boolean> ofI;

    @JsonProperty("OF_J")
    private final List<Boolean> ofJ;

    @JsonProperty("OF_RES")
    private final List<Boolean> ofRes;

    @JsonProperty("OLI")
    private final List<Boolean> oli;

    @JsonProperty("RES_HI")
    private final List<BigInteger> resHi;

    @JsonProperty("RES_LO")
    private final List<BigInteger> resLo;

    @JsonProperty("STAMP")
    private final List<BigInteger> stamp;

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
      this.accDelta0 = new ArrayList<>(length);
      this.accDelta1 = new ArrayList<>(length);
      this.accDelta2 = new ArrayList<>(length);
      this.accDelta3 = new ArrayList<>(length);
      this.accH0 = new ArrayList<>(length);
      this.accH1 = new ArrayList<>(length);
      this.accH2 = new ArrayList<>(length);
      this.accH3 = new ArrayList<>(length);
      this.accH4 = new ArrayList<>(length);
      this.accH5 = new ArrayList<>(length);
      this.accI0 = new ArrayList<>(length);
      this.accI1 = new ArrayList<>(length);
      this.accI2 = new ArrayList<>(length);
      this.accI3 = new ArrayList<>(length);
      this.accI4 = new ArrayList<>(length);
      this.accI5 = new ArrayList<>(length);
      this.accI6 = new ArrayList<>(length);
      this.accJ0 = new ArrayList<>(length);
      this.accJ1 = new ArrayList<>(length);
      this.accJ2 = new ArrayList<>(length);
      this.accJ3 = new ArrayList<>(length);
      this.accJ4 = new ArrayList<>(length);
      this.accJ5 = new ArrayList<>(length);
      this.accJ6 = new ArrayList<>(length);
      this.accJ7 = new ArrayList<>(length);
      this.accQ0 = new ArrayList<>(length);
      this.accQ1 = new ArrayList<>(length);
      this.accQ2 = new ArrayList<>(length);
      this.accQ3 = new ArrayList<>(length);
      this.accQ4 = new ArrayList<>(length);
      this.accQ5 = new ArrayList<>(length);
      this.accQ6 = new ArrayList<>(length);
      this.accQ7 = new ArrayList<>(length);
      this.accR0 = new ArrayList<>(length);
      this.accR1 = new ArrayList<>(length);
      this.accR2 = new ArrayList<>(length);
      this.accR3 = new ArrayList<>(length);
      this.arg1Hi = new ArrayList<>(length);
      this.arg1Lo = new ArrayList<>(length);
      this.arg2Hi = new ArrayList<>(length);
      this.arg2Lo = new ArrayList<>(length);
      this.arg3Hi = new ArrayList<>(length);
      this.arg3Lo = new ArrayList<>(length);
      this.bit1 = new ArrayList<>(length);
      this.bit2 = new ArrayList<>(length);
      this.bit3 = new ArrayList<>(length);
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
      this.byteDelta0 = new ArrayList<>(length);
      this.byteDelta1 = new ArrayList<>(length);
      this.byteDelta2 = new ArrayList<>(length);
      this.byteDelta3 = new ArrayList<>(length);
      this.byteH0 = new ArrayList<>(length);
      this.byteH1 = new ArrayList<>(length);
      this.byteH2 = new ArrayList<>(length);
      this.byteH3 = new ArrayList<>(length);
      this.byteH4 = new ArrayList<>(length);
      this.byteH5 = new ArrayList<>(length);
      this.byteI0 = new ArrayList<>(length);
      this.byteI1 = new ArrayList<>(length);
      this.byteI2 = new ArrayList<>(length);
      this.byteI3 = new ArrayList<>(length);
      this.byteI4 = new ArrayList<>(length);
      this.byteI5 = new ArrayList<>(length);
      this.byteI6 = new ArrayList<>(length);
      this.byteJ0 = new ArrayList<>(length);
      this.byteJ1 = new ArrayList<>(length);
      this.byteJ2 = new ArrayList<>(length);
      this.byteJ3 = new ArrayList<>(length);
      this.byteJ4 = new ArrayList<>(length);
      this.byteJ5 = new ArrayList<>(length);
      this.byteJ6 = new ArrayList<>(length);
      this.byteJ7 = new ArrayList<>(length);
      this.byteQ0 = new ArrayList<>(length);
      this.byteQ1 = new ArrayList<>(length);
      this.byteQ2 = new ArrayList<>(length);
      this.byteQ3 = new ArrayList<>(length);
      this.byteQ4 = new ArrayList<>(length);
      this.byteQ5 = new ArrayList<>(length);
      this.byteQ6 = new ArrayList<>(length);
      this.byteQ7 = new ArrayList<>(length);
      this.byteR0 = new ArrayList<>(length);
      this.byteR1 = new ArrayList<>(length);
      this.byteR2 = new ArrayList<>(length);
      this.byteR3 = new ArrayList<>(length);
      this.cmp = new ArrayList<>(length);
      this.ct = new ArrayList<>(length);
      this.inst = new ArrayList<>(length);
      this.ofH = new ArrayList<>(length);
      this.ofI = new ArrayList<>(length);
      this.ofJ = new ArrayList<>(length);
      this.ofRes = new ArrayList<>(length);
      this.oli = new ArrayList<>(length);
      this.resHi = new ArrayList<>(length);
      this.resLo = new ArrayList<>(length);
      this.stamp = new ArrayList<>(length);
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

    public TraceBuilder accDelta0(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("ACC_DELTA_0 already set");
      } else {
        filled.set(12);
      }

      accDelta0.add(b);

      return this;
    }

    public TraceBuilder accDelta1(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("ACC_DELTA_1 already set");
      } else {
        filled.set(13);
      }

      accDelta1.add(b);

      return this;
    }

    public TraceBuilder accDelta2(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("ACC_DELTA_2 already set");
      } else {
        filled.set(14);
      }

      accDelta2.add(b);

      return this;
    }

    public TraceBuilder accDelta3(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("ACC_DELTA_3 already set");
      } else {
        filled.set(15);
      }

      accDelta3.add(b);

      return this;
    }

    public TraceBuilder accH0(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("ACC_H_0 already set");
      } else {
        filled.set(16);
      }

      accH0.add(b);

      return this;
    }

    public TraceBuilder accH1(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("ACC_H_1 already set");
      } else {
        filled.set(17);
      }

      accH1.add(b);

      return this;
    }

    public TraceBuilder accH2(final BigInteger b) {
      if (filled.get(18)) {
        throw new IllegalStateException("ACC_H_2 already set");
      } else {
        filled.set(18);
      }

      accH2.add(b);

      return this;
    }

    public TraceBuilder accH3(final BigInteger b) {
      if (filled.get(19)) {
        throw new IllegalStateException("ACC_H_3 already set");
      } else {
        filled.set(19);
      }

      accH3.add(b);

      return this;
    }

    public TraceBuilder accH4(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("ACC_H_4 already set");
      } else {
        filled.set(20);
      }

      accH4.add(b);

      return this;
    }

    public TraceBuilder accH5(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("ACC_H_5 already set");
      } else {
        filled.set(21);
      }

      accH5.add(b);

      return this;
    }

    public TraceBuilder accI0(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("ACC_I_0 already set");
      } else {
        filled.set(22);
      }

      accI0.add(b);

      return this;
    }

    public TraceBuilder accI1(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("ACC_I_1 already set");
      } else {
        filled.set(23);
      }

      accI1.add(b);

      return this;
    }

    public TraceBuilder accI2(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("ACC_I_2 already set");
      } else {
        filled.set(24);
      }

      accI2.add(b);

      return this;
    }

    public TraceBuilder accI3(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ACC_I_3 already set");
      } else {
        filled.set(25);
      }

      accI3.add(b);

      return this;
    }

    public TraceBuilder accI4(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("ACC_I_4 already set");
      } else {
        filled.set(26);
      }

      accI4.add(b);

      return this;
    }

    public TraceBuilder accI5(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("ACC_I_5 already set");
      } else {
        filled.set(27);
      }

      accI5.add(b);

      return this;
    }

    public TraceBuilder accI6(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("ACC_I_6 already set");
      } else {
        filled.set(28);
      }

      accI6.add(b);

      return this;
    }

    public TraceBuilder accJ0(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("ACC_J_0 already set");
      } else {
        filled.set(29);
      }

      accJ0.add(b);

      return this;
    }

    public TraceBuilder accJ1(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("ACC_J_1 already set");
      } else {
        filled.set(30);
      }

      accJ1.add(b);

      return this;
    }

    public TraceBuilder accJ2(final BigInteger b) {
      if (filled.get(31)) {
        throw new IllegalStateException("ACC_J_2 already set");
      } else {
        filled.set(31);
      }

      accJ2.add(b);

      return this;
    }

    public TraceBuilder accJ3(final BigInteger b) {
      if (filled.get(32)) {
        throw new IllegalStateException("ACC_J_3 already set");
      } else {
        filled.set(32);
      }

      accJ3.add(b);

      return this;
    }

    public TraceBuilder accJ4(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("ACC_J_4 already set");
      } else {
        filled.set(33);
      }

      accJ4.add(b);

      return this;
    }

    public TraceBuilder accJ5(final BigInteger b) {
      if (filled.get(34)) {
        throw new IllegalStateException("ACC_J_5 already set");
      } else {
        filled.set(34);
      }

      accJ5.add(b);

      return this;
    }

    public TraceBuilder accJ6(final BigInteger b) {
      if (filled.get(35)) {
        throw new IllegalStateException("ACC_J_6 already set");
      } else {
        filled.set(35);
      }

      accJ6.add(b);

      return this;
    }

    public TraceBuilder accJ7(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("ACC_J_7 already set");
      } else {
        filled.set(36);
      }

      accJ7.add(b);

      return this;
    }

    public TraceBuilder accQ0(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("ACC_Q_0 already set");
      } else {
        filled.set(37);
      }

      accQ0.add(b);

      return this;
    }

    public TraceBuilder accQ1(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("ACC_Q_1 already set");
      } else {
        filled.set(38);
      }

      accQ1.add(b);

      return this;
    }

    public TraceBuilder accQ2(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("ACC_Q_2 already set");
      } else {
        filled.set(39);
      }

      accQ2.add(b);

      return this;
    }

    public TraceBuilder accQ3(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("ACC_Q_3 already set");
      } else {
        filled.set(40);
      }

      accQ3.add(b);

      return this;
    }

    public TraceBuilder accQ4(final BigInteger b) {
      if (filled.get(41)) {
        throw new IllegalStateException("ACC_Q_4 already set");
      } else {
        filled.set(41);
      }

      accQ4.add(b);

      return this;
    }

    public TraceBuilder accQ5(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("ACC_Q_5 already set");
      } else {
        filled.set(42);
      }

      accQ5.add(b);

      return this;
    }

    public TraceBuilder accQ6(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("ACC_Q_6 already set");
      } else {
        filled.set(43);
      }

      accQ6.add(b);

      return this;
    }

    public TraceBuilder accQ7(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("ACC_Q_7 already set");
      } else {
        filled.set(44);
      }

      accQ7.add(b);

      return this;
    }

    public TraceBuilder accR0(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("ACC_R_0 already set");
      } else {
        filled.set(45);
      }

      accR0.add(b);

      return this;
    }

    public TraceBuilder accR1(final BigInteger b) {
      if (filled.get(46)) {
        throw new IllegalStateException("ACC_R_1 already set");
      } else {
        filled.set(46);
      }

      accR1.add(b);

      return this;
    }

    public TraceBuilder accR2(final BigInteger b) {
      if (filled.get(47)) {
        throw new IllegalStateException("ACC_R_2 already set");
      } else {
        filled.set(47);
      }

      accR2.add(b);

      return this;
    }

    public TraceBuilder accR3(final BigInteger b) {
      if (filled.get(48)) {
        throw new IllegalStateException("ACC_R_3 already set");
      } else {
        filled.set(48);
      }

      accR3.add(b);

      return this;
    }

    public TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(49)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(49);
      }

      arg1Hi.add(b);

      return this;
    }

    public TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(50);
      }

      arg1Lo.add(b);

      return this;
    }

    public TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(51);
      }

      arg2Hi.add(b);

      return this;
    }

    public TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(52);
      }

      arg2Lo.add(b);

      return this;
    }

    public TraceBuilder arg3Hi(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("ARG_3_HI already set");
      } else {
        filled.set(53);
      }

      arg3Hi.add(b);

      return this;
    }

    public TraceBuilder arg3Lo(final BigInteger b) {
      if (filled.get(54)) {
        throw new IllegalStateException("ARG_3_LO already set");
      } else {
        filled.set(54);
      }

      arg3Lo.add(b);

      return this;
    }

    public TraceBuilder bit1(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(55);
      }

      bit1.add(b);

      return this;
    }

    public TraceBuilder bit2(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(56);
      }

      bit2.add(b);

      return this;
    }

    public TraceBuilder bit3(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(57);
      }

      bit3.add(b);

      return this;
    }

    public TraceBuilder byteA0(final UnsignedByte b) {
      if (filled.get(58)) {
        throw new IllegalStateException("BYTE_A_0 already set");
      } else {
        filled.set(58);
      }

      byteA0.add(b);

      return this;
    }

    public TraceBuilder byteA1(final UnsignedByte b) {
      if (filled.get(59)) {
        throw new IllegalStateException("BYTE_A_1 already set");
      } else {
        filled.set(59);
      }

      byteA1.add(b);

      return this;
    }

    public TraceBuilder byteA2(final UnsignedByte b) {
      if (filled.get(60)) {
        throw new IllegalStateException("BYTE_A_2 already set");
      } else {
        filled.set(60);
      }

      byteA2.add(b);

      return this;
    }

    public TraceBuilder byteA3(final UnsignedByte b) {
      if (filled.get(61)) {
        throw new IllegalStateException("BYTE_A_3 already set");
      } else {
        filled.set(61);
      }

      byteA3.add(b);

      return this;
    }

    public TraceBuilder byteB0(final UnsignedByte b) {
      if (filled.get(62)) {
        throw new IllegalStateException("BYTE_B_0 already set");
      } else {
        filled.set(62);
      }

      byteB0.add(b);

      return this;
    }

    public TraceBuilder byteB1(final UnsignedByte b) {
      if (filled.get(63)) {
        throw new IllegalStateException("BYTE_B_1 already set");
      } else {
        filled.set(63);
      }

      byteB1.add(b);

      return this;
    }

    public TraceBuilder byteB2(final UnsignedByte b) {
      if (filled.get(64)) {
        throw new IllegalStateException("BYTE_B_2 already set");
      } else {
        filled.set(64);
      }

      byteB2.add(b);

      return this;
    }

    public TraceBuilder byteB3(final UnsignedByte b) {
      if (filled.get(65)) {
        throw new IllegalStateException("BYTE_B_3 already set");
      } else {
        filled.set(65);
      }

      byteB3.add(b);

      return this;
    }

    public TraceBuilder byteC0(final UnsignedByte b) {
      if (filled.get(66)) {
        throw new IllegalStateException("BYTE_C_0 already set");
      } else {
        filled.set(66);
      }

      byteC0.add(b);

      return this;
    }

    public TraceBuilder byteC1(final UnsignedByte b) {
      if (filled.get(67)) {
        throw new IllegalStateException("BYTE_C_1 already set");
      } else {
        filled.set(67);
      }

      byteC1.add(b);

      return this;
    }

    public TraceBuilder byteC2(final UnsignedByte b) {
      if (filled.get(68)) {
        throw new IllegalStateException("BYTE_C_2 already set");
      } else {
        filled.set(68);
      }

      byteC2.add(b);

      return this;
    }

    public TraceBuilder byteC3(final UnsignedByte b) {
      if (filled.get(69)) {
        throw new IllegalStateException("BYTE_C_3 already set");
      } else {
        filled.set(69);
      }

      byteC3.add(b);

      return this;
    }

    public TraceBuilder byteDelta0(final UnsignedByte b) {
      if (filled.get(70)) {
        throw new IllegalStateException("BYTE_DELTA_0 already set");
      } else {
        filled.set(70);
      }

      byteDelta0.add(b);

      return this;
    }

    public TraceBuilder byteDelta1(final UnsignedByte b) {
      if (filled.get(71)) {
        throw new IllegalStateException("BYTE_DELTA_1 already set");
      } else {
        filled.set(71);
      }

      byteDelta1.add(b);

      return this;
    }

    public TraceBuilder byteDelta2(final UnsignedByte b) {
      if (filled.get(72)) {
        throw new IllegalStateException("BYTE_DELTA_2 already set");
      } else {
        filled.set(72);
      }

      byteDelta2.add(b);

      return this;
    }

    public TraceBuilder byteDelta3(final UnsignedByte b) {
      if (filled.get(73)) {
        throw new IllegalStateException("BYTE_DELTA_3 already set");
      } else {
        filled.set(73);
      }

      byteDelta3.add(b);

      return this;
    }

    public TraceBuilder byteH0(final UnsignedByte b) {
      if (filled.get(74)) {
        throw new IllegalStateException("BYTE_H_0 already set");
      } else {
        filled.set(74);
      }

      byteH0.add(b);

      return this;
    }

    public TraceBuilder byteH1(final UnsignedByte b) {
      if (filled.get(75)) {
        throw new IllegalStateException("BYTE_H_1 already set");
      } else {
        filled.set(75);
      }

      byteH1.add(b);

      return this;
    }

    public TraceBuilder byteH2(final UnsignedByte b) {
      if (filled.get(76)) {
        throw new IllegalStateException("BYTE_H_2 already set");
      } else {
        filled.set(76);
      }

      byteH2.add(b);

      return this;
    }

    public TraceBuilder byteH3(final UnsignedByte b) {
      if (filled.get(77)) {
        throw new IllegalStateException("BYTE_H_3 already set");
      } else {
        filled.set(77);
      }

      byteH3.add(b);

      return this;
    }

    public TraceBuilder byteH4(final UnsignedByte b) {
      if (filled.get(78)) {
        throw new IllegalStateException("BYTE_H_4 already set");
      } else {
        filled.set(78);
      }

      byteH4.add(b);

      return this;
    }

    public TraceBuilder byteH5(final UnsignedByte b) {
      if (filled.get(79)) {
        throw new IllegalStateException("BYTE_H_5 already set");
      } else {
        filled.set(79);
      }

      byteH5.add(b);

      return this;
    }

    public TraceBuilder byteI0(final UnsignedByte b) {
      if (filled.get(80)) {
        throw new IllegalStateException("BYTE_I_0 already set");
      } else {
        filled.set(80);
      }

      byteI0.add(b);

      return this;
    }

    public TraceBuilder byteI1(final UnsignedByte b) {
      if (filled.get(81)) {
        throw new IllegalStateException("BYTE_I_1 already set");
      } else {
        filled.set(81);
      }

      byteI1.add(b);

      return this;
    }

    public TraceBuilder byteI2(final UnsignedByte b) {
      if (filled.get(82)) {
        throw new IllegalStateException("BYTE_I_2 already set");
      } else {
        filled.set(82);
      }

      byteI2.add(b);

      return this;
    }

    public TraceBuilder byteI3(final UnsignedByte b) {
      if (filled.get(83)) {
        throw new IllegalStateException("BYTE_I_3 already set");
      } else {
        filled.set(83);
      }

      byteI3.add(b);

      return this;
    }

    public TraceBuilder byteI4(final UnsignedByte b) {
      if (filled.get(84)) {
        throw new IllegalStateException("BYTE_I_4 already set");
      } else {
        filled.set(84);
      }

      byteI4.add(b);

      return this;
    }

    public TraceBuilder byteI5(final UnsignedByte b) {
      if (filled.get(85)) {
        throw new IllegalStateException("BYTE_I_5 already set");
      } else {
        filled.set(85);
      }

      byteI5.add(b);

      return this;
    }

    public TraceBuilder byteI6(final UnsignedByte b) {
      if (filled.get(86)) {
        throw new IllegalStateException("BYTE_I_6 already set");
      } else {
        filled.set(86);
      }

      byteI6.add(b);

      return this;
    }

    public TraceBuilder byteJ0(final UnsignedByte b) {
      if (filled.get(87)) {
        throw new IllegalStateException("BYTE_J_0 already set");
      } else {
        filled.set(87);
      }

      byteJ0.add(b);

      return this;
    }

    public TraceBuilder byteJ1(final UnsignedByte b) {
      if (filled.get(88)) {
        throw new IllegalStateException("BYTE_J_1 already set");
      } else {
        filled.set(88);
      }

      byteJ1.add(b);

      return this;
    }

    public TraceBuilder byteJ2(final UnsignedByte b) {
      if (filled.get(89)) {
        throw new IllegalStateException("BYTE_J_2 already set");
      } else {
        filled.set(89);
      }

      byteJ2.add(b);

      return this;
    }

    public TraceBuilder byteJ3(final UnsignedByte b) {
      if (filled.get(90)) {
        throw new IllegalStateException("BYTE_J_3 already set");
      } else {
        filled.set(90);
      }

      byteJ3.add(b);

      return this;
    }

    public TraceBuilder byteJ4(final UnsignedByte b) {
      if (filled.get(91)) {
        throw new IllegalStateException("BYTE_J_4 already set");
      } else {
        filled.set(91);
      }

      byteJ4.add(b);

      return this;
    }

    public TraceBuilder byteJ5(final UnsignedByte b) {
      if (filled.get(92)) {
        throw new IllegalStateException("BYTE_J_5 already set");
      } else {
        filled.set(92);
      }

      byteJ5.add(b);

      return this;
    }

    public TraceBuilder byteJ6(final UnsignedByte b) {
      if (filled.get(93)) {
        throw new IllegalStateException("BYTE_J_6 already set");
      } else {
        filled.set(93);
      }

      byteJ6.add(b);

      return this;
    }

    public TraceBuilder byteJ7(final UnsignedByte b) {
      if (filled.get(94)) {
        throw new IllegalStateException("BYTE_J_7 already set");
      } else {
        filled.set(94);
      }

      byteJ7.add(b);

      return this;
    }

    public TraceBuilder byteQ0(final UnsignedByte b) {
      if (filled.get(95)) {
        throw new IllegalStateException("BYTE_Q_0 already set");
      } else {
        filled.set(95);
      }

      byteQ0.add(b);

      return this;
    }

    public TraceBuilder byteQ1(final UnsignedByte b) {
      if (filled.get(96)) {
        throw new IllegalStateException("BYTE_Q_1 already set");
      } else {
        filled.set(96);
      }

      byteQ1.add(b);

      return this;
    }

    public TraceBuilder byteQ2(final UnsignedByte b) {
      if (filled.get(97)) {
        throw new IllegalStateException("BYTE_Q_2 already set");
      } else {
        filled.set(97);
      }

      byteQ2.add(b);

      return this;
    }

    public TraceBuilder byteQ3(final UnsignedByte b) {
      if (filled.get(98)) {
        throw new IllegalStateException("BYTE_Q_3 already set");
      } else {
        filled.set(98);
      }

      byteQ3.add(b);

      return this;
    }

    public TraceBuilder byteQ4(final UnsignedByte b) {
      if (filled.get(99)) {
        throw new IllegalStateException("BYTE_Q_4 already set");
      } else {
        filled.set(99);
      }

      byteQ4.add(b);

      return this;
    }

    public TraceBuilder byteQ5(final UnsignedByte b) {
      if (filled.get(100)) {
        throw new IllegalStateException("BYTE_Q_5 already set");
      } else {
        filled.set(100);
      }

      byteQ5.add(b);

      return this;
    }

    public TraceBuilder byteQ6(final UnsignedByte b) {
      if (filled.get(101)) {
        throw new IllegalStateException("BYTE_Q_6 already set");
      } else {
        filled.set(101);
      }

      byteQ6.add(b);

      return this;
    }

    public TraceBuilder byteQ7(final UnsignedByte b) {
      if (filled.get(102)) {
        throw new IllegalStateException("BYTE_Q_7 already set");
      } else {
        filled.set(102);
      }

      byteQ7.add(b);

      return this;
    }

    public TraceBuilder byteR0(final UnsignedByte b) {
      if (filled.get(103)) {
        throw new IllegalStateException("BYTE_R_0 already set");
      } else {
        filled.set(103);
      }

      byteR0.add(b);

      return this;
    }

    public TraceBuilder byteR1(final UnsignedByte b) {
      if (filled.get(104)) {
        throw new IllegalStateException("BYTE_R_1 already set");
      } else {
        filled.set(104);
      }

      byteR1.add(b);

      return this;
    }

    public TraceBuilder byteR2(final UnsignedByte b) {
      if (filled.get(105)) {
        throw new IllegalStateException("BYTE_R_2 already set");
      } else {
        filled.set(105);
      }

      byteR2.add(b);

      return this;
    }

    public TraceBuilder byteR3(final UnsignedByte b) {
      if (filled.get(106)) {
        throw new IllegalStateException("BYTE_R_3 already set");
      } else {
        filled.set(106);
      }

      byteR3.add(b);

      return this;
    }

    public TraceBuilder cmp(final Boolean b) {
      if (filled.get(107)) {
        throw new IllegalStateException("CMP already set");
      } else {
        filled.set(107);
      }

      cmp.add(b);

      return this;
    }

    public TraceBuilder ct(final BigInteger b) {
      if (filled.get(108)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(108);
      }

      ct.add(b);

      return this;
    }

    public TraceBuilder inst(final BigInteger b) {
      if (filled.get(109)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(109);
      }

      inst.add(b);

      return this;
    }

    public TraceBuilder ofH(final Boolean b) {
      if (filled.get(110)) {
        throw new IllegalStateException("OF_H already set");
      } else {
        filled.set(110);
      }

      ofH.add(b);

      return this;
    }

    public TraceBuilder ofI(final Boolean b) {
      if (filled.get(111)) {
        throw new IllegalStateException("OF_I already set");
      } else {
        filled.set(111);
      }

      ofI.add(b);

      return this;
    }

    public TraceBuilder ofJ(final Boolean b) {
      if (filled.get(112)) {
        throw new IllegalStateException("OF_J already set");
      } else {
        filled.set(112);
      }

      ofJ.add(b);

      return this;
    }

    public TraceBuilder ofRes(final Boolean b) {
      if (filled.get(113)) {
        throw new IllegalStateException("OF_RES already set");
      } else {
        filled.set(113);
      }

      ofRes.add(b);

      return this;
    }

    public TraceBuilder oli(final Boolean b) {
      if (filled.get(114)) {
        throw new IllegalStateException("OLI already set");
      } else {
        filled.set(114);
      }

      oli.add(b);

      return this;
    }

    public TraceBuilder resHi(final BigInteger b) {
      if (filled.get(115)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(115);
      }

      resHi.add(b);

      return this;
    }

    public TraceBuilder resLo(final BigInteger b) {
      if (filled.get(116)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(116);
      }

      resLo.add(b);

      return this;
    }

    public TraceBuilder stamp(final BigInteger b) {
      if (filled.get(117)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(117);
      }

      stamp.add(b);

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
        throw new IllegalStateException("ACC_DELTA_0 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("ACC_DELTA_1 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("ACC_DELTA_2 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("ACC_DELTA_3 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("ACC_H_0 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("ACC_H_1 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("ACC_H_2 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("ACC_H_3 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("ACC_H_4 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("ACC_H_5 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("ACC_I_0 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("ACC_I_1 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("ACC_I_2 has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ACC_I_3 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("ACC_I_4 has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("ACC_I_5 has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("ACC_I_6 has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("ACC_J_0 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("ACC_J_1 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("ACC_J_2 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("ACC_J_3 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("ACC_J_4 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("ACC_J_5 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("ACC_J_6 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("ACC_J_7 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("ACC_Q_0 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("ACC_Q_1 has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("ACC_Q_2 has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("ACC_Q_3 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("ACC_Q_4 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("ACC_Q_5 has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("ACC_Q_6 has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("ACC_Q_7 has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("ACC_R_0 has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("ACC_R_1 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("ACC_R_2 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("ACC_R_3 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("ARG_3_HI has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("ARG_3_LO has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("BYTE_A_0 has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("BYTE_A_1 has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("BYTE_A_2 has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException("BYTE_A_3 has not been filled");
      }

      if (!filled.get(62)) {
        throw new IllegalStateException("BYTE_B_0 has not been filled");
      }

      if (!filled.get(63)) {
        throw new IllegalStateException("BYTE_B_1 has not been filled");
      }

      if (!filled.get(64)) {
        throw new IllegalStateException("BYTE_B_2 has not been filled");
      }

      if (!filled.get(65)) {
        throw new IllegalStateException("BYTE_B_3 has not been filled");
      }

      if (!filled.get(66)) {
        throw new IllegalStateException("BYTE_C_0 has not been filled");
      }

      if (!filled.get(67)) {
        throw new IllegalStateException("BYTE_C_1 has not been filled");
      }

      if (!filled.get(68)) {
        throw new IllegalStateException("BYTE_C_2 has not been filled");
      }

      if (!filled.get(69)) {
        throw new IllegalStateException("BYTE_C_3 has not been filled");
      }

      if (!filled.get(70)) {
        throw new IllegalStateException("BYTE_DELTA_0 has not been filled");
      }

      if (!filled.get(71)) {
        throw new IllegalStateException("BYTE_DELTA_1 has not been filled");
      }

      if (!filled.get(72)) {
        throw new IllegalStateException("BYTE_DELTA_2 has not been filled");
      }

      if (!filled.get(73)) {
        throw new IllegalStateException("BYTE_DELTA_3 has not been filled");
      }

      if (!filled.get(74)) {
        throw new IllegalStateException("BYTE_H_0 has not been filled");
      }

      if (!filled.get(75)) {
        throw new IllegalStateException("BYTE_H_1 has not been filled");
      }

      if (!filled.get(76)) {
        throw new IllegalStateException("BYTE_H_2 has not been filled");
      }

      if (!filled.get(77)) {
        throw new IllegalStateException("BYTE_H_3 has not been filled");
      }

      if (!filled.get(78)) {
        throw new IllegalStateException("BYTE_H_4 has not been filled");
      }

      if (!filled.get(79)) {
        throw new IllegalStateException("BYTE_H_5 has not been filled");
      }

      if (!filled.get(80)) {
        throw new IllegalStateException("BYTE_I_0 has not been filled");
      }

      if (!filled.get(81)) {
        throw new IllegalStateException("BYTE_I_1 has not been filled");
      }

      if (!filled.get(82)) {
        throw new IllegalStateException("BYTE_I_2 has not been filled");
      }

      if (!filled.get(83)) {
        throw new IllegalStateException("BYTE_I_3 has not been filled");
      }

      if (!filled.get(84)) {
        throw new IllegalStateException("BYTE_I_4 has not been filled");
      }

      if (!filled.get(85)) {
        throw new IllegalStateException("BYTE_I_5 has not been filled");
      }

      if (!filled.get(86)) {
        throw new IllegalStateException("BYTE_I_6 has not been filled");
      }

      if (!filled.get(87)) {
        throw new IllegalStateException("BYTE_J_0 has not been filled");
      }

      if (!filled.get(88)) {
        throw new IllegalStateException("BYTE_J_1 has not been filled");
      }

      if (!filled.get(89)) {
        throw new IllegalStateException("BYTE_J_2 has not been filled");
      }

      if (!filled.get(90)) {
        throw new IllegalStateException("BYTE_J_3 has not been filled");
      }

      if (!filled.get(91)) {
        throw new IllegalStateException("BYTE_J_4 has not been filled");
      }

      if (!filled.get(92)) {
        throw new IllegalStateException("BYTE_J_5 has not been filled");
      }

      if (!filled.get(93)) {
        throw new IllegalStateException("BYTE_J_6 has not been filled");
      }

      if (!filled.get(94)) {
        throw new IllegalStateException("BYTE_J_7 has not been filled");
      }

      if (!filled.get(95)) {
        throw new IllegalStateException("BYTE_Q_0 has not been filled");
      }

      if (!filled.get(96)) {
        throw new IllegalStateException("BYTE_Q_1 has not been filled");
      }

      if (!filled.get(97)) {
        throw new IllegalStateException("BYTE_Q_2 has not been filled");
      }

      if (!filled.get(98)) {
        throw new IllegalStateException("BYTE_Q_3 has not been filled");
      }

      if (!filled.get(99)) {
        throw new IllegalStateException("BYTE_Q_4 has not been filled");
      }

      if (!filled.get(100)) {
        throw new IllegalStateException("BYTE_Q_5 has not been filled");
      }

      if (!filled.get(101)) {
        throw new IllegalStateException("BYTE_Q_6 has not been filled");
      }

      if (!filled.get(102)) {
        throw new IllegalStateException("BYTE_Q_7 has not been filled");
      }

      if (!filled.get(103)) {
        throw new IllegalStateException("BYTE_R_0 has not been filled");
      }

      if (!filled.get(104)) {
        throw new IllegalStateException("BYTE_R_1 has not been filled");
      }

      if (!filled.get(105)) {
        throw new IllegalStateException("BYTE_R_2 has not been filled");
      }

      if (!filled.get(106)) {
        throw new IllegalStateException("BYTE_R_3 has not been filled");
      }

      if (!filled.get(107)) {
        throw new IllegalStateException("CMP has not been filled");
      }

      if (!filled.get(108)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(109)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(110)) {
        throw new IllegalStateException("OF_H has not been filled");
      }

      if (!filled.get(111)) {
        throw new IllegalStateException("OF_I has not been filled");
      }

      if (!filled.get(112)) {
        throw new IllegalStateException("OF_J has not been filled");
      }

      if (!filled.get(113)) {
        throw new IllegalStateException("OF_RES has not been filled");
      }

      if (!filled.get(114)) {
        throw new IllegalStateException("OLI has not been filled");
      }

      if (!filled.get(115)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(116)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(117)) {
        throw new IllegalStateException("STAMP has not been filled");
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
        accDelta0.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        accDelta1.add(BigInteger.ZERO);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        accDelta2.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        accDelta3.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        accH0.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        accH1.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        accH2.add(BigInteger.ZERO);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        accH3.add(BigInteger.ZERO);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        accH4.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        accH5.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        accI0.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        accI1.add(BigInteger.ZERO);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        accI2.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        accI3.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        accI4.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        accI5.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        accI6.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        accJ0.add(BigInteger.ZERO);
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        accJ1.add(BigInteger.ZERO);
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        accJ2.add(BigInteger.ZERO);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        accJ3.add(BigInteger.ZERO);
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        accJ4.add(BigInteger.ZERO);
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        accJ5.add(BigInteger.ZERO);
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        accJ6.add(BigInteger.ZERO);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        accJ7.add(BigInteger.ZERO);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        accQ0.add(BigInteger.ZERO);
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        accQ1.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        accQ2.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        accQ3.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        accQ4.add(BigInteger.ZERO);
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        accQ5.add(BigInteger.ZERO);
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        accQ6.add(BigInteger.ZERO);
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        accQ7.add(BigInteger.ZERO);
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        accR0.add(BigInteger.ZERO);
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        accR1.add(BigInteger.ZERO);
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        accR2.add(BigInteger.ZERO);
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        accR3.add(BigInteger.ZERO);
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        arg1Hi.add(BigInteger.ZERO);
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        arg1Lo.add(BigInteger.ZERO);
        this.filled.set(50);
      }
      if (!filled.get(51)) {
        arg2Hi.add(BigInteger.ZERO);
        this.filled.set(51);
      }
      if (!filled.get(52)) {
        arg2Lo.add(BigInteger.ZERO);
        this.filled.set(52);
      }
      if (!filled.get(53)) {
        arg3Hi.add(BigInteger.ZERO);
        this.filled.set(53);
      }
      if (!filled.get(54)) {
        arg3Lo.add(BigInteger.ZERO);
        this.filled.set(54);
      }
      if (!filled.get(55)) {
        bit1.add(false);
        this.filled.set(55);
      }
      if (!filled.get(56)) {
        bit2.add(false);
        this.filled.set(56);
      }
      if (!filled.get(57)) {
        bit3.add(false);
        this.filled.set(57);
      }
      if (!filled.get(58)) {
        byteA0.add(UnsignedByte.of(0));
        this.filled.set(58);
      }
      if (!filled.get(59)) {
        byteA1.add(UnsignedByte.of(0));
        this.filled.set(59);
      }
      if (!filled.get(60)) {
        byteA2.add(UnsignedByte.of(0));
        this.filled.set(60);
      }
      if (!filled.get(61)) {
        byteA3.add(UnsignedByte.of(0));
        this.filled.set(61);
      }
      if (!filled.get(62)) {
        byteB0.add(UnsignedByte.of(0));
        this.filled.set(62);
      }
      if (!filled.get(63)) {
        byteB1.add(UnsignedByte.of(0));
        this.filled.set(63);
      }
      if (!filled.get(64)) {
        byteB2.add(UnsignedByte.of(0));
        this.filled.set(64);
      }
      if (!filled.get(65)) {
        byteB3.add(UnsignedByte.of(0));
        this.filled.set(65);
      }
      if (!filled.get(66)) {
        byteC0.add(UnsignedByte.of(0));
        this.filled.set(66);
      }
      if (!filled.get(67)) {
        byteC1.add(UnsignedByte.of(0));
        this.filled.set(67);
      }
      if (!filled.get(68)) {
        byteC2.add(UnsignedByte.of(0));
        this.filled.set(68);
      }
      if (!filled.get(69)) {
        byteC3.add(UnsignedByte.of(0));
        this.filled.set(69);
      }
      if (!filled.get(70)) {
        byteDelta0.add(UnsignedByte.of(0));
        this.filled.set(70);
      }
      if (!filled.get(71)) {
        byteDelta1.add(UnsignedByte.of(0));
        this.filled.set(71);
      }
      if (!filled.get(72)) {
        byteDelta2.add(UnsignedByte.of(0));
        this.filled.set(72);
      }
      if (!filled.get(73)) {
        byteDelta3.add(UnsignedByte.of(0));
        this.filled.set(73);
      }
      if (!filled.get(74)) {
        byteH0.add(UnsignedByte.of(0));
        this.filled.set(74);
      }
      if (!filled.get(75)) {
        byteH1.add(UnsignedByte.of(0));
        this.filled.set(75);
      }
      if (!filled.get(76)) {
        byteH2.add(UnsignedByte.of(0));
        this.filled.set(76);
      }
      if (!filled.get(77)) {
        byteH3.add(UnsignedByte.of(0));
        this.filled.set(77);
      }
      if (!filled.get(78)) {
        byteH4.add(UnsignedByte.of(0));
        this.filled.set(78);
      }
      if (!filled.get(79)) {
        byteH5.add(UnsignedByte.of(0));
        this.filled.set(79);
      }
      if (!filled.get(80)) {
        byteI0.add(UnsignedByte.of(0));
        this.filled.set(80);
      }
      if (!filled.get(81)) {
        byteI1.add(UnsignedByte.of(0));
        this.filled.set(81);
      }
      if (!filled.get(82)) {
        byteI2.add(UnsignedByte.of(0));
        this.filled.set(82);
      }
      if (!filled.get(83)) {
        byteI3.add(UnsignedByte.of(0));
        this.filled.set(83);
      }
      if (!filled.get(84)) {
        byteI4.add(UnsignedByte.of(0));
        this.filled.set(84);
      }
      if (!filled.get(85)) {
        byteI5.add(UnsignedByte.of(0));
        this.filled.set(85);
      }
      if (!filled.get(86)) {
        byteI6.add(UnsignedByte.of(0));
        this.filled.set(86);
      }
      if (!filled.get(87)) {
        byteJ0.add(UnsignedByte.of(0));
        this.filled.set(87);
      }
      if (!filled.get(88)) {
        byteJ1.add(UnsignedByte.of(0));
        this.filled.set(88);
      }
      if (!filled.get(89)) {
        byteJ2.add(UnsignedByte.of(0));
        this.filled.set(89);
      }
      if (!filled.get(90)) {
        byteJ3.add(UnsignedByte.of(0));
        this.filled.set(90);
      }
      if (!filled.get(91)) {
        byteJ4.add(UnsignedByte.of(0));
        this.filled.set(91);
      }
      if (!filled.get(92)) {
        byteJ5.add(UnsignedByte.of(0));
        this.filled.set(92);
      }
      if (!filled.get(93)) {
        byteJ6.add(UnsignedByte.of(0));
        this.filled.set(93);
      }
      if (!filled.get(94)) {
        byteJ7.add(UnsignedByte.of(0));
        this.filled.set(94);
      }
      if (!filled.get(95)) {
        byteQ0.add(UnsignedByte.of(0));
        this.filled.set(95);
      }
      if (!filled.get(96)) {
        byteQ1.add(UnsignedByte.of(0));
        this.filled.set(96);
      }
      if (!filled.get(97)) {
        byteQ2.add(UnsignedByte.of(0));
        this.filled.set(97);
      }
      if (!filled.get(98)) {
        byteQ3.add(UnsignedByte.of(0));
        this.filled.set(98);
      }
      if (!filled.get(99)) {
        byteQ4.add(UnsignedByte.of(0));
        this.filled.set(99);
      }
      if (!filled.get(100)) {
        byteQ5.add(UnsignedByte.of(0));
        this.filled.set(100);
      }
      if (!filled.get(101)) {
        byteQ6.add(UnsignedByte.of(0));
        this.filled.set(101);
      }
      if (!filled.get(102)) {
        byteQ7.add(UnsignedByte.of(0));
        this.filled.set(102);
      }
      if (!filled.get(103)) {
        byteR0.add(UnsignedByte.of(0));
        this.filled.set(103);
      }
      if (!filled.get(104)) {
        byteR1.add(UnsignedByte.of(0));
        this.filled.set(104);
      }
      if (!filled.get(105)) {
        byteR2.add(UnsignedByte.of(0));
        this.filled.set(105);
      }
      if (!filled.get(106)) {
        byteR3.add(UnsignedByte.of(0));
        this.filled.set(106);
      }
      if (!filled.get(107)) {
        cmp.add(false);
        this.filled.set(107);
      }
      if (!filled.get(108)) {
        ct.add(BigInteger.ZERO);
        this.filled.set(108);
      }
      if (!filled.get(109)) {
        inst.add(BigInteger.ZERO);
        this.filled.set(109);
      }
      if (!filled.get(110)) {
        ofH.add(false);
        this.filled.set(110);
      }
      if (!filled.get(111)) {
        ofI.add(false);
        this.filled.set(111);
      }
      if (!filled.get(112)) {
        ofJ.add(false);
        this.filled.set(112);
      }
      if (!filled.get(113)) {
        ofRes.add(false);
        this.filled.set(113);
      }
      if (!filled.get(114)) {
        oli.add(false);
        this.filled.set(114);
      }
      if (!filled.get(115)) {
        resHi.add(BigInteger.ZERO);
        this.filled.set(115);
      }
      if (!filled.get(116)) {
        resLo.add(BigInteger.ZERO);
        this.filled.set(116);
      }
      if (!filled.get(117)) {
        stamp.add(BigInteger.ZERO);
        this.filled.set(117);
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
