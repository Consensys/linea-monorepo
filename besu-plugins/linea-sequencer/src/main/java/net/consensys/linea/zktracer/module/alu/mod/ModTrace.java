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
package net.consensys.linea.zktracer.module.alu.mod;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({"Trace", "Stamp"})
@SuppressWarnings("unused")
public record ModTrace(@JsonProperty("Trace") Trace trace, @JsonProperty("Stamp") int stamp) {

  @JsonPropertyOrder({
      "ACC_1_2",
      "ACC_1_3",
      "ACC_2_2",
      "ACC_2_3",
      "ACC_B_0",
      "ACC_B_1",
      "ACC_B_2",
      "ACC_B_3",
      "ACC_DELTA_0",
      "ACC_DELTA_1",
      "ACC_DELTA_2",
      "ACC_DELTA_3",
      "ACC_H_0",
      "ACC_H_1",
      "ACC_H_2",
      "ACC_Q_0",
      "ACC_Q_1",
      "ACC_Q_2",
      "ACC_Q_3",
      "ACC_R_0",
      "ACC_R_1",
      "ACC_R_2",
      "ACC_R_3",
      "ARG_1_HI",
      "ARG_1_LO",
      "ARG_2_HI",
      "ARG_2_LO",
      "BYTE_1_2",
      "BYTE_1_3",
      "BYTE_2_2",
      "BYTE_2_3",
      "BYTE_B_0",
      "BYTE_B_1",
      "BYTE_B_2",
      "BYTE_B_3",
      "BYTE_DELTA_0",
      "BYTE_DELTA_1",
      "BYTE_DELTA_2",
      "BYTE_DELTA_3",
      "BYTE_H_0",
      "BYTE_H_1",
      "BYTE_H_2",
      "BYTE_Q_0",
      "BYTE_Q_1",
      "BYTE_Q_2",
      "BYTE_Q_3",
      "BYTE_R_0",
      "BYTE_R_1",
      "BYTE_R_2",
      "BYTE_R_3",
      "CMP_1",
      "CMP_2",
      "CT",
      "DEC_OUTPUT",
      "DEC_SIGNED",
      "INST",
      "MSB_1",
      "MSB_2",
      "OLI",
      "RES_HI",
      "RES_LO",
      "STAMP",
  })
  @SuppressWarnings("unused")
  public record Trace(
      @JsonProperty("ACC_1_2") List<BigInteger> ACC_1_2,
      @JsonProperty("ACC_1_3") List<BigInteger> ACC_1_3,
      @JsonProperty("ACC_2_2") List<BigInteger> ACC_2_2,
      @JsonProperty("ACC_2_3") List<BigInteger> ACC_2_3,
      @JsonProperty("ACC_B_0") List<BigInteger> ACC_B_0,
      @JsonProperty("ACC_B_1") List<BigInteger> ACC_B_1,
      @JsonProperty("ACC_B_2") List<BigInteger> ACC_B_2,
      @JsonProperty("ACC_B_3") List<BigInteger> ACC_B_3,
      @JsonProperty("ACC_DELTA_0") List<BigInteger> ACC_DELTA_0,
      @JsonProperty("ACC_DELTA_1") List<BigInteger> ACC_DELTA_1,
      @JsonProperty("ACC_DELTA_2") List<BigInteger> ACC_DELTA_2,
      @JsonProperty("ACC_DELTA_3") List<BigInteger> ACC_DELTA_3,
      @JsonProperty("ACC_H_0") List<BigInteger> ACC_H_0,
      @JsonProperty("ACC_H_1") List<BigInteger> ACC_H_1,
      @JsonProperty("ACC_H_2") List<BigInteger> ACC_H_2,
      @JsonProperty("ACC_Q_0") List<BigInteger> ACC_Q_0,
      @JsonProperty("ACC_Q_1") List<BigInteger> ACC_Q_1,
      @JsonProperty("ACC_Q_2") List<BigInteger> ACC_Q_2,
      @JsonProperty("ACC_Q_3") List<BigInteger> ACC_Q_3,
      @JsonProperty("ACC_R_0") List<BigInteger> ACC_R_0,
      @JsonProperty("ACC_R_1") List<BigInteger> ACC_R_1,
      @JsonProperty("ACC_R_2") List<BigInteger> ACC_R_2,
      @JsonProperty("ACC_R_3") List<BigInteger> ACC_R_3,
      @JsonProperty("ARG_1_HI") List<BigInteger> ARG_1_HI,
      @JsonProperty("ARG_1_LO") List<BigInteger> ARG_1_LO,
      @JsonProperty("ARG_2_HI") List<BigInteger> ARG_2_HI,
      @JsonProperty("ARG_2_LO") List<BigInteger> ARG_2_LO,
      @JsonProperty("BYTE_1_2") List<UnsignedByte> BYTE_1_2,
      @JsonProperty("BYTE_1_3") List<UnsignedByte> BYTE_1_3,
      @JsonProperty("BYTE_2_2") List<UnsignedByte> BYTE_2_2,
      @JsonProperty("BYTE_2_3") List<UnsignedByte> BYTE_2_3,
      @JsonProperty("BYTE_B_0") List<UnsignedByte> BYTE_B_0,
      @JsonProperty("BYTE_B_1") List<UnsignedByte> BYTE_B_1,
      @JsonProperty("BYTE_B_2") List<UnsignedByte> BYTE_B_2,
      @JsonProperty("BYTE_B_3") List<UnsignedByte> BYTE_B_3,
      @JsonProperty("BYTE_DELTA_0") List<UnsignedByte> BYTE_DELTA_0,
      @JsonProperty("BYTE_DELTA_1") List<UnsignedByte> BYTE_DELTA_1,
      @JsonProperty("BYTE_DELTA_2") List<UnsignedByte> BYTE_DELTA_2,
      @JsonProperty("BYTE_DELTA_3") List<UnsignedByte> BYTE_DELTA_3,
      @JsonProperty("BYTE_H_0") List<UnsignedByte> BYTE_H_0,
      @JsonProperty("BYTE_H_1") List<UnsignedByte> BYTE_H_1,
      @JsonProperty("BYTE_H_2") List<UnsignedByte> BYTE_H_2,
      @JsonProperty("BYTE_Q_0") List<UnsignedByte> BYTE_Q_0,
      @JsonProperty("BYTE_Q_1") List<UnsignedByte> BYTE_Q_1,
      @JsonProperty("BYTE_Q_2") List<UnsignedByte> BYTE_Q_2,
      @JsonProperty("BYTE_Q_3") List<UnsignedByte> BYTE_Q_3,
      @JsonProperty("BYTE_R_0") List<UnsignedByte> BYTE_R_0,
      @JsonProperty("BYTE_R_1") List<UnsignedByte> BYTE_R_1,
      @JsonProperty("BYTE_R_2") List<UnsignedByte> BYTE_R_2,
      @JsonProperty("BYTE_R_3") List<UnsignedByte> BYTE_R_3,
      @JsonProperty("CMP_1") List<Boolean> CMP_1,
      @JsonProperty("CMP_2") List<Boolean> CMP_2,
      @JsonProperty("CT") List<Integer> CT,
      @JsonProperty("DEC_OUTPUT") List<Boolean> DEC_OUTPUT,
      @JsonProperty("DEC_SIGNED") List<Boolean> DEC_SIGNED,
      @JsonProperty("INST") List<UnsignedByte> INST,
      @JsonProperty("MSB_1") List<Boolean> MSB_1,
      @JsonProperty("MSB_2") List<Boolean> MSB_2,
      @JsonProperty("OLI") List<Boolean> OLI,
      @JsonProperty("RES_HI") List<BigInteger> RES_HI,
      @JsonProperty("RES_LO") List<BigInteger> RES_LO,
      @JsonProperty("STAMP") List<Integer> STAMP) {

    public static class Builder {

      private final List<BigInteger> acc_1_2 = new ArrayList<>();
      private final List<BigInteger> acc_1_3 = new ArrayList<>();
      private final List<BigInteger> acc_2_2 = new ArrayList<>();
      private final List<BigInteger> acc_2_3 = new ArrayList<>();
      private final List<BigInteger> acc_B_0 = new ArrayList<>();
      private final List<BigInteger> acc_B_1 = new ArrayList<>();
      private final List<BigInteger> acc_B_2 = new ArrayList<>();
      private final List<BigInteger> acc_B_3 = new ArrayList<>();
      private final List<BigInteger> accDelta_0 = new ArrayList<>();
      private final List<BigInteger> accDelta_1 = new ArrayList<>();
      private final List<BigInteger> accDelta_2 = new ArrayList<>();
      private final List<BigInteger> accDelta_3 = new ArrayList<>();
      private final List<BigInteger> acc_H_0 = new ArrayList<>();
      private final List<BigInteger> acc_H_1 = new ArrayList<>();
      private final List<BigInteger> acc_H_2 = new ArrayList<>();
      private final List<BigInteger> acc_Q_0 = new ArrayList<>();
      private final List<BigInteger> acc_Q_1 = new ArrayList<>();
      private final List<BigInteger> acc_Q_2 = new ArrayList<>();
      private final List<BigInteger> acc_Q_3 = new ArrayList<>();
      private final List<BigInteger> acc_R_0 = new ArrayList<>();
      private final List<BigInteger> acc_R_1 = new ArrayList<>();
      private final List<BigInteger> acc_R_2 = new ArrayList<>();
      private final List<BigInteger> acc_R_3 = new ArrayList<>();
      private final List<BigInteger> arg_1_HI = new ArrayList<>();
      private final List<BigInteger> arg_1_LO = new ArrayList<>();
      private final List<BigInteger> arg_2_HI = new ArrayList<>();
      private final List<BigInteger> arg_2_LO = new ArrayList<>();
      private final List<UnsignedByte> byte_1_2 = new ArrayList<>();
      private final List<UnsignedByte> byte_1_3 = new ArrayList<>();
      private final List<UnsignedByte> byte_2_2 = new ArrayList<>();
      private final List<UnsignedByte> byte_2_3 = new ArrayList<>();
      private final List<UnsignedByte> byte_B_0 = new ArrayList<>();
      private final List<UnsignedByte> byte_B_1 = new ArrayList<>();
      private final List<UnsignedByte> byte_B_2 = new ArrayList<>();
      private final List<UnsignedByte> byte_B_3 = new ArrayList<>();
      private final List<UnsignedByte> byteDelta_0 = new ArrayList<>();
      private final List<UnsignedByte> byteDelta_1 = new ArrayList<>();
      private final List<UnsignedByte> byteDelta_2 = new ArrayList<>();
      private final List<UnsignedByte> byteDelta_3 = new ArrayList<>();
      private final List<UnsignedByte> byte_H_0 = new ArrayList<>();
      private final List<UnsignedByte> byte_H_1 = new ArrayList<>();
      private final List<UnsignedByte> byte_H_2 = new ArrayList<>();
      private final List<UnsignedByte> byte_Q_0 = new ArrayList<>();
      private final List<UnsignedByte> byte_Q_1 = new ArrayList<>();
      private final List<UnsignedByte> byte_Q_2 = new ArrayList<>();
      private final List<UnsignedByte> byte_Q_3 = new ArrayList<>();
      private final List<UnsignedByte> byte_R_0 = new ArrayList<>();
      private final List<UnsignedByte> byte_R_1 = new ArrayList<>();
      private final List<UnsignedByte> byte_R_2 = new ArrayList<>();
      private final List<UnsignedByte> byte_R_3 = new ArrayList<>();
      private final List<Boolean> cmp1 = new ArrayList<>();
      private final List<Boolean> cmp2 = new ArrayList<>();
      private final List<Integer> ct = new ArrayList<>();
      private final List<Boolean> decOutput = new ArrayList<>();
      private final List<Boolean> decSigned = new ArrayList<>();
      private final List<UnsignedByte> inst = new ArrayList<>();
      private final List<Boolean> msb1 = new ArrayList<>();
      private final List<Boolean> msb2 = new ArrayList<>();
      private final List<Boolean> oli = new ArrayList<>();
      private final List<BigInteger> resHi = new ArrayList<>();
      private final List<BigInteger> resLo = new ArrayList<>();
      private final List<Integer> modStamp = new ArrayList<>();
      private int stamp = 0;

      private Builder() {
      }

      public static Builder newInstance() {
        return new Builder();
      }

      public Builder appendAcc_1_2(final BigInteger b) {
        acc_1_2.add(b);
        return this;
      }

      public Builder appendAcc_1_3(final BigInteger b) {
        acc_1_3.add(b);
        return this;
      }

      public Builder appendAcc_2_2(final BigInteger b) {
        acc_2_2.add(b);
        return this;
      }

      public Builder appendAcc_2_3(final BigInteger b) {
        acc_2_3.add(b);
        return this;
      }

      public Builder appendAcc_B_0(final BigInteger b) {
        acc_B_0.add(b);
        return this;
      }

      public Builder appendAcc_B_1(final BigInteger b) {
        acc_B_1.add(b);
        return this;
      }

      public Builder appendAcc_B_2(final BigInteger b) {
        acc_B_2.add(b);
        return this;
      }

      public Builder appendAcc_B_3(final BigInteger b) {
        acc_B_3.add(b);
        return this;
      }

      public Builder appendAccDelta_0(final BigInteger b) {
        accDelta_0.add(b);
        return this;
      }

      public Builder appendAccDelta_1(final BigInteger b) {
        accDelta_1.add(b);
        return this;
      }

      public Builder appendAccDelta_2(final BigInteger b) {
        accDelta_2.add(b);
        return this;
      }

      public Builder appendAccDelta_3(final BigInteger b) {
        accDelta_3.add(b);
        return this;
      }

      public Builder appendAcc_H_0(final BigInteger b) {
        acc_H_0.add(b);
        return this;
      }

      public Builder appendAcc_H_1(final BigInteger b) {
        acc_H_1.add(b);
        return this;
      }

      public Builder appendAcc_H_2(final BigInteger b) {
        acc_H_2.add(b);
        return this;
      }

      public Builder appendAcc_Q_0(final BigInteger b) {
        acc_Q_0.add(b);
        return this;
      }

      public Builder appendAcc_Q_1(final BigInteger b) {
        acc_Q_1.add(b);
        return this;
      }

      public Builder appendAcc_Q_2(final BigInteger b) {
        acc_Q_2.add(b);
        return this;
      }

      public Builder appendAcc_Q_3(final BigInteger b) {
        acc_Q_3.add(b);
        return this;
      }

      public Builder appendAcc_R_0(final BigInteger b) {
        acc_R_0.add(b);
        return this;
      }

      public Builder appendAcc_R_1(final BigInteger b) {
        acc_R_1.add(b);
        return this;
      }

      public Builder appendAcc_R_2(final BigInteger b) {
        acc_R_2.add(b);
        return this;
      }

      public Builder appendAcc_R_3(final BigInteger b) {
        acc_R_3.add(b);
        return this;
      }

      public Builder appendArg_1_HI(final BigInteger b) {
        arg_1_HI.add(b);
        return this;
      }

      public Builder appendArg_1_LO(final BigInteger b) {
        arg_1_LO.add(b);
        return this;
      }

      public Builder appendArg_2_HI(final BigInteger b) {
        arg_2_HI.add(b);
        return this;
      }

      public Builder appendArg_2_LO(final BigInteger b) {
        arg_2_LO.add(b);
        return this;
      }

      public Builder appendByte_1_2(final UnsignedByte b) {
        byte_1_2.add(b);
        return this;
      }

      public Builder appendByte_1_3(final UnsignedByte b) {
        byte_1_3.add(b);
        return this;
      }

      public Builder appendByte_2_2(final UnsignedByte b) {
        byte_2_2.add(b);
        return this;
      }

      public Builder appendByte_2_3(final UnsignedByte b) {
        byte_2_3.add(b);
        return this;
      }

      public Builder appendByte_B_0(final UnsignedByte b) {
        byte_B_0.add(b);
        return this;
      }

      public Builder appendByte_B_1(final UnsignedByte b) {
        byte_B_1.add(b);
        return this;
      }

      public Builder appendByte_B_2(final UnsignedByte b) {
        byte_B_2.add(b);
        return this;
      }

      public Builder appendByte_B_3(final UnsignedByte b) {
        byte_B_3.add(b);
        return this;
      }

      public Builder appendByteDelta_0(final UnsignedByte b) {
        byteDelta_0.add(b);
        return this;
      }

      public Builder appendByteDelta_1(final UnsignedByte b) {
        byteDelta_1.add(b);
        return this;
      }

      public Builder appendByteDelta_2(final UnsignedByte b) {
        byteDelta_2.add(b);
        return this;
      }

      public Builder appendByteDelta_3(final UnsignedByte b) {
        byteDelta_3.add(b);
        return this;
      }

      public Builder appendByte_H_0(final UnsignedByte b) {
        byte_H_0.add(b);
        return this;
      }

      public Builder appendByte_H_1(final UnsignedByte b) {
        byte_H_1.add(b);
        return this;
      }

      public Builder appendByte_H_2(final UnsignedByte b) {
        byte_H_2.add(b);
        return this;
      }

      public Builder appendByte_Q_0(final UnsignedByte b) {
        byte_Q_0.add(b);
        return this;
      }

      public Builder appendByte_Q_1(final UnsignedByte b) {
        byte_Q_1.add(b);
        return this;
      }

      public Builder appendByte_Q_2(final UnsignedByte b) {
        byte_Q_2.add(b);
        return this;
      }

      public Builder appendByte_Q_3(final UnsignedByte b) {
        byte_Q_3.add(b);
        return this;
      }

      public Builder appendByte_R_0(final UnsignedByte b) {
        byte_R_0.add(b);
        return this;
      }

      public Builder appendByte_R_1(final UnsignedByte b) {
        byte_R_1.add(b);
        return this;
      }

      public Builder appendByte_R_2(final UnsignedByte b) {
        byte_R_2.add(b);
        return this;
      }

      public Builder appendByte_R_3(final UnsignedByte b) {
        byte_R_3.add(b);
        return this;
      }

      public Builder appendCmp1(final Boolean b) {
        cmp1.add(b);
        return this;
      }

      public Builder appendCmp2(final Boolean b) {
        cmp2.add(b);
        return this;
      }

      public Builder appendCt(final Integer b) {
        ct.add(b);
        return this;
      }

      public Builder appendDecOutput(final Boolean b) {
        decOutput.add(b);
        return this;
      }

      public Builder appendArg1Hi(final BigInteger b) {
        arg_1_HI.add(b);
        return this;
      }

      public Builder appendArg1Lo(final BigInteger b) {
        arg_1_LO.add(b);
        return this;
      }

      public Builder appendArg2Hi(final BigInteger b) {
        arg_2_HI.add(b);
        return this;
      }

      public Builder appendArg2Lo(final BigInteger b) {
        arg_2_LO.add(b);
        return this;
      }

      public Builder appendDecSigned(final Boolean b) {
        decSigned.add(b);
        return this;
      }

      public Builder appendInst(final UnsignedByte b) {
        inst.add(b);
        return this;
      }

      public Builder appendMsb1(final Boolean b) {
        msb1.add(b);
        return this;
      }

      public Builder appendMsb2(final Boolean b) {
        msb2.add(b);
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

      public Builder appendModStamp(final Integer b) {
        modStamp.add(b);
        return this;
      }


      public ModTrace build() {
        return new ModTrace(
            new Trace(
                acc_1_2,
                acc_1_3,
                acc_2_2,
                acc_2_3,
                acc_B_0,
                acc_B_1,
                acc_B_2,
                acc_B_3,
                accDelta_0,
                accDelta_1,
                accDelta_2,
                accDelta_3,
                acc_H_0,
                acc_H_1,
                acc_H_2,
                acc_Q_0,
                acc_Q_1,
                acc_Q_2,
                acc_Q_3,
                acc_R_0,
                acc_R_1,
                acc_R_2,
                acc_R_3,
                arg_1_HI,
                arg_1_LO,
                arg_2_HI,
                arg_2_LO,
                byte_1_2,
                byte_1_3,
                byte_2_2,
                byte_2_3,
                byte_B_0,
                byte_B_1,
                byte_B_2,
                byte_B_3,
                byteDelta_0,
                byteDelta_1,
                byteDelta_2,
                byteDelta_3,
                byte_H_0,
                byte_H_1,
                byte_H_2,
                byte_Q_0,
                byte_Q_1,
                byte_Q_2,
                byte_Q_3,
                byte_R_0,
                byte_R_1,
                byte_R_2,
                byte_R_3,
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
                modStamp),
            stamp);
      }
    }
  }
}
