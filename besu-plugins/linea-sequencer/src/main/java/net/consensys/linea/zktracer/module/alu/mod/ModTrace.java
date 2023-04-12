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

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
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
      @JsonProperty("STAMP") List<Integer> MOD_STAMP) {

    public static class Builder {

      private final List<BigInteger> acc1_2 = new ArrayList<>();
      private final List<BigInteger> acc1_3 = new ArrayList<>();
      private final List<BigInteger> acc2_2 = new ArrayList<>();
      private final List<BigInteger> acc2_3 = new ArrayList<>();
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
      private final List<UnsignedByte> byte1_2 = new ArrayList<>();
      private final List<UnsignedByte> byte1_3 = new ArrayList<>();
      private final List<UnsignedByte> byte2_2 = new ArrayList<>();
      private final List<UnsignedByte> byte2_3 = new ArrayList<>();
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

      private Builder() {}

      public static Builder newInstance() {
        return new Builder();
      }

      public Builder appendAcc1_2(final BigInteger b) {
        acc1_2.add(b);
        return this;
      }

      public Builder appendAcc1_3(final BigInteger b) {
        acc1_3.add(b);
        return this;
      }

      public Builder appendAcc2_2(final BigInteger b) {
        acc2_2.add(b);
        return this;
      }

      public Builder appendAcc2_3(final BigInteger b) {
        acc2_3.add(b);
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

      public Builder appendArg_1_HI(final BigInteger b) {
        arg1Hi.add(b);
        return this;
      }

      public Builder appendArg_1_LO(final BigInteger b) {
        arg1Lo.add(b);
        return this;
      }

      public Builder appendArg_2_HI(final BigInteger b) {
        arg2Hi.add(b);
        return this;
      }

      public Builder appendArg_2_LO(final BigInteger b) {
        arg2Lo.add(b);
        return this;
      }

      public Builder appendByte1_2(final UnsignedByte b) {
        byte1_2.add(b);
        return this;
      }

      public Builder appendByte1_3(final UnsignedByte b) {
        byte1_3.add(b);
        return this;
      }

      public Builder appendByte2_2(final UnsignedByte b) {
        byte2_2.add(b);
        return this;
      }

      public Builder appendByte2_3(final UnsignedByte b) {
        byte2_3.add(b);
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

      public Builder setStamp(final int stamp) {
        this.stamp = stamp;
        return this;
      }

      public ModTrace build() {
        return new ModTrace(
            new Trace(
                acc1_2,
                acc1_3,
                acc2_2,
                acc2_3,
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
                byte1_2,
                byte1_3,
                byte2_2,
                byte2_3,
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
                modStamp),
            stamp);
      }
    }
  }
}
