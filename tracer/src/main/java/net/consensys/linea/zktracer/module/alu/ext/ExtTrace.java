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
package net.consensys.linea.zktracer.module.alu.ext;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({"Trace", "Stamp"})
@SuppressWarnings("unused")
public record ExtTrace(@JsonProperty("Trace") Trace trace, @JsonProperty("Stamp") int stamp) {

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
    "ACC_DELTA_0",
    "ACC_DELTA_1",
    "ACC_DELTA_2",
    "ACC_DELTA_3",
    "ACC_H_0",
    "ACC_H_1",
    "ACC_H_2",
    "ACC_H_3",
    "ACC_H_4",
    "ACC_H_5",
    "ACC_I_0",
    "ACC_I_1",
    "ACC_I_2",
    "ACC_I_3",
    "ACC_I_4",
    "ACC_I_5",
    "ACC_I_6",
    "ACC_J_0",
    "ACC_J_1",
    "ACC_J_2",
    "ACC_J_3",
    "ACC_J_4",
    "ACC_J_5",
    "ACC_J_6",
    "ACC_J_7",
    "ACC_Q_0",
    "ACC_Q_1",
    "ACC_Q_2",
    "ACC_Q_3",
    "ACC_Q_4",
    "ACC_Q_5",
    "ACC_Q_6",
    "ACC_Q_7",
    "ACC_R_0",
    "ACC_R_1",
    "ACC_R_2",
    "ACC_R_3",
    "ARG_1_HI",
    "ARG_1_LO",
    "ARG_2_HI",
    "ARG_2_LO",
    "ARG_3_HI",
    "ARG_3_LO",
    "BIT_1",
    "BIT_2",
    "BIT_3",
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
    "BYTE_DELTA_0",
    "BYTE_DELTA_1",
    "BYTE_DELTA_2",
    "BYTE_DELTA_3",
    "BYTE_H_0",
    "BYTE_H_1",
    "BYTE_H_2",
    "BYTE_H_3",
    "BYTE_H_4",
    "BYTE_H_5",
    "BYTE_I_0",
    "BYTE_I_1",
    "BYTE_I_2",
    "BYTE_I_3",
    "BYTE_I_4",
    "BYTE_I_5",
    "BYTE_I_6",
    "BYTE_J_0",
    "BYTE_J_1",
    "BYTE_J_2",
    "BYTE_J_3",
    "BYTE_J_4",
    "BYTE_J_5",
    "BYTE_J_6",
    "BYTE_J_7",
    "BYTE_Q_0",
    "BYTE_Q_1",
    "BYTE_Q_2",
    "BYTE_Q_3",
    "BYTE_Q_4",
    "BYTE_Q_5",
    "BYTE_Q_6",
    "BYTE_Q_7",
    "BYTE_R_0",
    "BYTE_R_1",
    "BYTE_R_2",
    "BYTE_R_3",
    "CMP",
    "CT",
    "INST",
    "OF_H",
    "OF_I",
    "OF_J",
    "OF_RES",
    "OLI",
    "RES_HI",
    "RES_LO",
    "STAMP",
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
      @JsonProperty("ACC_DELTA_0") List<BigInteger> ACC_DELTA_0,
      @JsonProperty("ACC_DELTA_1") List<BigInteger> ACC_DELTA_1,
      @JsonProperty("ACC_DELTA_2") List<BigInteger> ACC_DELTA_2,
      @JsonProperty("ACC_DELTA_3") List<BigInteger> ACC_DELTA_3,
      @JsonProperty("ACC_H_0") List<BigInteger> ACC_H_0,
      @JsonProperty("ACC_H_1") List<BigInteger> ACC_H_1,
      @JsonProperty("ACC_H_2") List<BigInteger> ACC_H_2,
      @JsonProperty("ACC_H_3") List<BigInteger> ACC_H_3,
      @JsonProperty("ACC_H_4") List<BigInteger> ACC_H_4,
      @JsonProperty("ACC_H_5") List<BigInteger> ACC_H_5,
      @JsonProperty("ACC_I_0") List<BigInteger> ACC_I_0,
      @JsonProperty("ACC_I_1") List<BigInteger> ACC_I_1,
      @JsonProperty("ACC_I_2") List<BigInteger> ACC_I_2,
      @JsonProperty("ACC_I_3") List<BigInteger> ACC_I_3,
      @JsonProperty("ACC_I_4") List<BigInteger> ACC_I_4,
      @JsonProperty("ACC_I_5") List<BigInteger> ACC_I_5,
      @JsonProperty("ACC_I_6") List<BigInteger> ACC_I_6,
      @JsonProperty("ACC_J_0") List<BigInteger> ACC_J_0,
      @JsonProperty("ACC_J_1") List<BigInteger> ACC_J_1,
      @JsonProperty("ACC_J_2") List<BigInteger> ACC_J_2,
      @JsonProperty("ACC_J_3") List<BigInteger> ACC_J_3,
      @JsonProperty("ACC_J_4") List<BigInteger> ACC_J_4,
      @JsonProperty("ACC_J_5") List<BigInteger> ACC_J_5,
      @JsonProperty("ACC_J_6") List<BigInteger> ACC_J_6,
      @JsonProperty("ACC_J_7") List<BigInteger> ACC_J_7,
      @JsonProperty("ACC_Q_0") List<BigInteger> ACC_Q_0,
      @JsonProperty("ACC_Q_1") List<BigInteger> ACC_Q_1,
      @JsonProperty("ACC_Q_2") List<BigInteger> ACC_Q_2,
      @JsonProperty("ACC_Q_3") List<BigInteger> ACC_Q_3,
      @JsonProperty("ACC_Q_4") List<BigInteger> ACC_Q_4,
      @JsonProperty("ACC_Q_5") List<BigInteger> ACC_Q_5,
      @JsonProperty("ACC_Q_6") List<BigInteger> ACC_Q_6,
      @JsonProperty("ACC_Q_7") List<BigInteger> ACC_Q_7,
      @JsonProperty("ACC_R_0") List<BigInteger> ACC_R_0,
      @JsonProperty("ACC_R_1") List<BigInteger> ACC_R_1,
      @JsonProperty("ACC_R_2") List<BigInteger> ACC_R_2,
      @JsonProperty("ACC_R_3") List<BigInteger> ACC_R_3,
      @JsonProperty("ARG_1_HI") List<BigInteger> ARG_1_HI,
      @JsonProperty("ARG_1_LO") List<BigInteger> ARG_1_LO,
      @JsonProperty("ARG_2_HI") List<BigInteger> ARG_2_HI,
      @JsonProperty("ARG_2_LO") List<BigInteger> ARG_2_LO,
      @JsonProperty("ARG_3_HI") List<BigInteger> ARG_3_HI,
      @JsonProperty("ARG_3_LO") List<BigInteger> ARG_3_LO,
      @JsonProperty("BIT_1") List<Boolean> BIT_1,
      @JsonProperty("BIT_2") List<Boolean> BIT_2,
      @JsonProperty("BIT_3") List<Boolean> BIT_3,
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
      @JsonProperty("BYTE_DELTA_0") List<UnsignedByte> BYTE_DELTA_0,
      @JsonProperty("BYTE_DELTA_1") List<UnsignedByte> BYTE_DELTA_1,
      @JsonProperty("BYTE_DELTA_2") List<UnsignedByte> BYTE_DELTA_2,
      @JsonProperty("BYTE_DELTA_3") List<UnsignedByte> BYTE_DELTA_3,
      @JsonProperty("BYTE_H_0") List<UnsignedByte> BYTE_H_0,
      @JsonProperty("BYTE_H_1") List<UnsignedByte> BYTE_H_1,
      @JsonProperty("BYTE_H_2") List<UnsignedByte> BYTE_H_2,
      @JsonProperty("BYTE_H_3") List<UnsignedByte> BYTE_H_3,
      @JsonProperty("BYTE_H_4") List<UnsignedByte> BYTE_H_4,
      @JsonProperty("BYTE_H_5") List<UnsignedByte> BYTE_H_5,
      @JsonProperty("BYTE_I_0") List<UnsignedByte> BYTE_I_0,
      @JsonProperty("BYTE_I_1") List<UnsignedByte> BYTE_I_1,
      @JsonProperty("BYTE_I_2") List<UnsignedByte> BYTE_I_2,
      @JsonProperty("BYTE_I_3") List<UnsignedByte> BYTE_I_3,
      @JsonProperty("BYTE_I_4") List<UnsignedByte> BYTE_I_4,
      @JsonProperty("BYTE_I_5") List<UnsignedByte> BYTE_I_5,
      @JsonProperty("BYTE_I_6") List<UnsignedByte> BYTE_I_6,
      @JsonProperty("BYTE_J_0") List<UnsignedByte> BYTE_J_0,
      @JsonProperty("BYTE_J_1") List<UnsignedByte> BYTE_J_1,
      @JsonProperty("BYTE_J_2") List<UnsignedByte> BYTE_J_2,
      @JsonProperty("BYTE_J_3") List<UnsignedByte> BYTE_J_3,
      @JsonProperty("BYTE_J_4") List<UnsignedByte> BYTE_J_4,
      @JsonProperty("BYTE_J_5") List<UnsignedByte> BYTE_J_5,
      @JsonProperty("BYTE_J_6") List<UnsignedByte> BYTE_J_6,
      @JsonProperty("BYTE_J_7") List<UnsignedByte> BYTE_J_7,
      @JsonProperty("BYTE_Q_0") List<UnsignedByte> BYTE_Q_0,
      @JsonProperty("BYTE_Q_1") List<UnsignedByte> BYTE_Q_1,
      @JsonProperty("BYTE_Q_2") List<UnsignedByte> BYTE_Q_2,
      @JsonProperty("BYTE_Q_3") List<UnsignedByte> BYTE_Q_3,
      @JsonProperty("BYTE_Q_4") List<UnsignedByte> BYTE_Q_4,
      @JsonProperty("BYTE_Q_5") List<UnsignedByte> BYTE_Q_5,
      @JsonProperty("BYTE_Q_6") List<UnsignedByte> BYTE_Q_6,
      @JsonProperty("BYTE_Q_7") List<UnsignedByte> BYTE_Q_7,
      @JsonProperty("BYTE_R_0") List<UnsignedByte> BYTE_R_0,
      @JsonProperty("BYTE_R_1") List<UnsignedByte> BYTE_R_1,
      @JsonProperty("BYTE_R_2") List<UnsignedByte> BYTE_R_2,
      @JsonProperty("BYTE_R_3") List<UnsignedByte> BYTE_R_3,
      @JsonProperty("CMP") List<Boolean> CMP,
      @JsonProperty("CT") List<Integer> CT,
      @JsonProperty("INST") List<UnsignedByte> INST,
      @JsonProperty("OF_H") List<Boolean> OF_H,
      @JsonProperty("OF_I") List<Boolean> OF_I,
      @JsonProperty("OF_J") List<Boolean> OF_J,
      @JsonProperty("OF_RES") List<Boolean> OF_RES,
      @JsonProperty("OLI") List<Boolean> OLI,
      @JsonProperty("RES_HI") List<BigInteger> RES_HI,
      @JsonProperty("RES_LO") List<BigInteger> RES_LO,
      @JsonProperty("STAMP") List<Integer> EXT_STAMP) {

    public static class Builder {
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
      private final List<Integer> ct = new ArrayList<>();
      private final List<UnsignedByte> inst = new ArrayList<>();
      private final List<Boolean> ofH = new ArrayList<>();
      private final List<Boolean> ofI = new ArrayList<>();
      private final List<Boolean> ofJ = new ArrayList<>();
      private final List<Boolean> ofRes = new ArrayList<>();
      private final List<Boolean> oli = new ArrayList<>();
      private final List<BigInteger> resHi = new ArrayList<>();
      private final List<BigInteger> resLo = new ArrayList<>();

      private final List<Integer> extStamp = new ArrayList<>();

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

      public Builder appendAccDelta0(final BigInteger b) {
        accDelta0.add(b);
        return this;
      }

      public Builder appendAccDelta1(final BigInteger b) {
        accDelta1.add(b);
        return this;
      }

      public Builder appendAccDelta2(final BigInteger b) {
        accDelta2.add(b);
        return this;
      }

      public Builder appendAccDelta3(final BigInteger b) {
        accDelta3.add(b);
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

      public Builder appendAccH4(final BigInteger b) {
        accH4.add(b);
        return this;
      }

      public Builder appendAccH5(final BigInteger b) {
        accH5.add(b);
        return this;
      }

      public Builder appendAccI0(final BigInteger b) {
        accI0.add(b);
        return this;
      }

      public Builder appendAccI1(final BigInteger b) {
        accI1.add(b);
        return this;
      }

      public Builder appendAccI2(final BigInteger b) {
        accI2.add(b);
        return this;
      }

      public Builder appendAccI3(final BigInteger b) {
        accI3.add(b);
        return this;
      }

      public Builder appendAccI4(final BigInteger b) {
        accI4.add(b);
        return this;
      }

      public Builder appendAccI5(final BigInteger b) {
        accI5.add(b);
        return this;
      }

      public Builder appendAccI6(final BigInteger b) {
        accI6.add(b);
        return this;
      }

      public Builder appendAccJ0(final BigInteger b) {
        accJ0.add(b);
        return this;
      }

      public Builder appendAccJ1(final BigInteger b) {
        accJ1.add(b);
        return this;
      }

      public Builder appendAccJ2(final BigInteger b) {
        accJ2.add(b);
        return this;
      }

      public Builder appendAccJ3(final BigInteger b) {
        accJ3.add(b);
        return this;
      }

      public Builder appendAccJ4(final BigInteger b) {
        accJ4.add(b);
        return this;
      }

      public Builder appendAccJ5(final BigInteger b) {
        accJ5.add(b);
        return this;
      }

      public Builder appendAccJ6(final BigInteger b) {
        accJ6.add(b);
        return this;
      }

      public Builder appendAccJ7(final BigInteger b) {
        accJ7.add(b);
        return this;
      }

      public Builder appendAccQ0(final BigInteger b) {
        accQ0.add(b);
        return this;
      }

      public Builder appendAccQ1(final BigInteger b) {
        accQ1.add(b);
        return this;
      }

      public Builder appendAccQ2(final BigInteger b) {
        accQ2.add(b);
        return this;
      }

      public Builder appendAccQ3(final BigInteger b) {
        accQ3.add(b);
        return this;
      }

      public Builder appendAccQ4(final BigInteger b) {
        accQ4.add(b);
        return this;
      }

      public Builder appendAccQ5(final BigInteger b) {
        accQ5.add(b);
        return this;
      }

      public Builder appendAccQ6(final BigInteger b) {
        accQ6.add(b);
        return this;
      }

      public Builder appendAccQ7(final BigInteger b) {
        accQ7.add(b);
        return this;
      }

      public Builder appendAccR0(final BigInteger b) {
        accR0.add(b);
        return this;
      }

      public Builder appendAccR1(final BigInteger b) {
        accR1.add(b);
        return this;
      }

      public Builder appendAccR2(final BigInteger b) {
        accR2.add(b);
        return this;
      }

      public Builder appendAccR3(final BigInteger b) {
        accR3.add(b);
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

      public Builder appendArg3Hi(final BigInteger b) {
        arg3Hi.add(b);
        return this;
      }

      public Builder appendArg3Lo(final BigInteger b) {
        arg3Lo.add(b);
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

      public Builder appendByteDelta0(final UnsignedByte b) {
        byteDelta0.add(b);
        return this;
      }

      public Builder appendByteDelta1(final UnsignedByte b) {
        byteDelta1.add(b);
        return this;
      }

      public Builder appendByteDelta2(final UnsignedByte b) {
        byteDelta2.add(b);
        return this;
      }

      public Builder appendByteDelta3(final UnsignedByte b) {
        byteDelta3.add(b);
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

      public Builder appendByteH4(final UnsignedByte b) {
        byteH4.add(b);
        return this;
      }

      public Builder appendByteH5(final UnsignedByte b) {
        byteH5.add(b);
        return this;
      }

      public Builder appendByteI0(final UnsignedByte b) {
        byteI0.add(b);
        return this;
      }

      public Builder appendByteI1(final UnsignedByte b) {
        byteI1.add(b);
        return this;
      }

      public Builder appendByteI2(final UnsignedByte b) {
        byteI2.add(b);
        return this;
      }

      public Builder appendByteI3(final UnsignedByte b) {
        byteI3.add(b);
        return this;
      }

      public Builder appendByteI4(final UnsignedByte b) {
        byteI4.add(b);
        return this;
      }

      public Builder appendByteI5(final UnsignedByte b) {
        byteI5.add(b);
        return this;
      }

      public Builder appendByteI6(final UnsignedByte b) {
        byteI6.add(b);
        return this;
      }

      public Builder appendByteJ0(final UnsignedByte b) {
        byteJ0.add(b);
        return this;
      }

      public Builder appendByteJ1(final UnsignedByte b) {
        byteJ1.add(b);
        return this;
      }

      public Builder appendByteJ2(final UnsignedByte b) {
        byteJ2.add(b);
        return this;
      }

      public Builder appendByteJ3(final UnsignedByte b) {
        byteJ3.add(b);
        return this;
      }

      public Builder appendByteJ4(final UnsignedByte b) {
        byteJ4.add(b);
        return this;
      }

      public Builder appendByteJ5(final UnsignedByte b) {
        byteJ5.add(b);
        return this;
      }

      public Builder appendByteJ6(final UnsignedByte b) {
        byteJ6.add(b);
        return this;
      }

      public Builder appendByteJ7(final UnsignedByte b) {
        byteJ7.add(b);
        return this;
      }

      public Builder appendByteQ0(final UnsignedByte b) {
        byteQ0.add(b);
        return this;
      }

      public Builder appendByteQ1(final UnsignedByte b) {
        byteQ1.add(b);
        return this;
      }

      public Builder appendByteQ2(final UnsignedByte b) {
        byteQ2.add(b);
        return this;
      }

      public Builder appendByteQ3(final UnsignedByte b) {
        byteQ3.add(b);
        return this;
      }

      public Builder appendByteQ4(final UnsignedByte b) {
        byteQ4.add(b);
        return this;
      }

      public Builder appendByteQ5(final UnsignedByte b) {
        byteQ5.add(b);
        return this;
      }

      public Builder appendByteQ6(final UnsignedByte b) {
        byteQ6.add(b);
        return this;
      }

      public Builder appendByteQ7(final UnsignedByte b) {
        byteQ7.add(b);
        return this;
      }

      public Builder appendByteR0(final UnsignedByte b) {
        byteR0.add(b);
        return this;
      }

      public Builder appendByteR1(final UnsignedByte b) {
        byteR1.add(b);
        return this;
      }

      public Builder appendByteR2(final UnsignedByte b) {
        byteR2.add(b);
        return this;
      }

      public Builder appendByteR3(final UnsignedByte b) {
        byteR3.add(b);
        return this;
      }

      public Builder appendCmp(final boolean b) {
        cmp.add(b);
        return this;
      }

      public Builder appendCt(final Integer b) {
        ct.add(b);
        return this;
      }

      public Builder appendInst(final UnsignedByte b) {
        inst.add(b);
        return this;
      }

      public Builder appendOfH(final Boolean b) {
        ofH.add(b);
        return this;
      }

      public Builder appendOfI(final Boolean b) {
        ofI.add(b);
        return this;
      }

      public Builder appendOfJ(final Boolean b) {
        ofJ.add(b);
        return this;
      }

      public Builder appendOfRes(final Boolean b) {
        ofRes.add(b);
        return this;
      }

      public Builder appendOli(final Boolean b) {
        oli.add(b);
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

      public Builder appendStamp(final Integer b) {
        extStamp.add(b);
        return this;
      }

      public Builder setStamp(final int stamp) {
        this.stamp = stamp;
        return this;
      }

      public ExtTrace build() {
        return new ExtTrace(
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
                extStamp),
            stamp);
      }
    }
  }
}
