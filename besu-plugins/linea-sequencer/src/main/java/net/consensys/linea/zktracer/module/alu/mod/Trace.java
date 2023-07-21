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
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import lombok.Builder;
import lombok.Singular;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

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
@Builder
record Trace(
    @Singular @JsonProperty("ACC_1_2") List<BigInteger> acc12Args,
    @Singular @JsonProperty("ACC_1_3") List<BigInteger> acc13Args,
    @Singular @JsonProperty("ACC_2_2") List<BigInteger> acc22Args,
    @Singular @JsonProperty("ACC_2_3") List<BigInteger> acc23Args,
    @Singular @JsonProperty("ACC_B_0") List<BigInteger> accB0Args,
    @Singular @JsonProperty("ACC_B_1") List<BigInteger> accB1Args,
    @Singular @JsonProperty("ACC_B_2") List<BigInteger> accB2Args,
    @Singular @JsonProperty("ACC_B_3") List<BigInteger> accB3Args,
    @Singular @JsonProperty("ACC_DELTA_0") List<BigInteger> accDelta0Args,
    @Singular @JsonProperty("ACC_DELTA_1") List<BigInteger> accDelta1Args,
    @Singular @JsonProperty("ACC_DELTA_2") List<BigInteger> accDelta2Args,
    @Singular @JsonProperty("ACC_DELTA_3") List<BigInteger> accDelta3Args,
    @Singular @JsonProperty("ACC_H_0") List<BigInteger> accH0Args,
    @Singular @JsonProperty("ACC_H_1") List<BigInteger> accH1Args,
    @Singular @JsonProperty("ACC_H_2") List<BigInteger> accH2Args,
    @Singular @JsonProperty("ACC_Q_0") List<BigInteger> accQ0Args,
    @Singular @JsonProperty("ACC_Q_1") List<BigInteger> accQ1Args,
    @Singular @JsonProperty("ACC_Q_2") List<BigInteger> accQ2Args,
    @Singular @JsonProperty("ACC_Q_3") List<BigInteger> accQ3Args,
    @Singular @JsonProperty("ACC_R_0") List<BigInteger> accR0Args,
    @Singular @JsonProperty("ACC_R_1") List<BigInteger> accR1Args,
    @Singular @JsonProperty("ACC_R_2") List<BigInteger> accR2Args,
    @Singular @JsonProperty("ACC_R_3") List<BigInteger> accR3Args,
    @Singular @JsonProperty("ARG_1_HI") List<BigInteger> arg1HiArgs,
    @Singular @JsonProperty("ARG_1_LO") List<BigInteger> arg1LoArgs,
    @Singular @JsonProperty("ARG_2_HI") List<BigInteger> arg2HiArgs,
    @Singular @JsonProperty("ARG_2_LO") List<BigInteger> arg2LoArgs,
    @Singular @JsonProperty("BYTE_1_2") List<UnsignedByte> byte12Args,
    @Singular @JsonProperty("BYTE_1_3") List<UnsignedByte> byte13Args,
    @Singular @JsonProperty("BYTE_2_2") List<UnsignedByte> byte22Args,
    @Singular @JsonProperty("BYTE_2_3") List<UnsignedByte> byte23Args,
    @Singular @JsonProperty("BYTE_B_0") List<UnsignedByte> byteB0Args,
    @Singular @JsonProperty("BYTE_B_1") List<UnsignedByte> byteB1Args,
    @Singular @JsonProperty("BYTE_B_2") List<UnsignedByte> byteB2Args,
    @Singular @JsonProperty("BYTE_B_3") List<UnsignedByte> byteB3Args,
    @Singular @JsonProperty("BYTE_DELTA_0") List<UnsignedByte> byteDelta0Args,
    @Singular @JsonProperty("BYTE_DELTA_1") List<UnsignedByte> byteDelta1Args,
    @Singular @JsonProperty("BYTE_DELTA_2") List<UnsignedByte> byteDelta2Args,
    @Singular @JsonProperty("BYTE_DELTA_3") List<UnsignedByte> byteDelta3Args,
    @Singular @JsonProperty("BYTE_H_0") List<UnsignedByte> byteH0Args,
    @Singular @JsonProperty("BYTE_H_1") List<UnsignedByte> byteH1Args,
    @Singular @JsonProperty("BYTE_H_2") List<UnsignedByte> byteH2Args,
    @Singular @JsonProperty("BYTE_Q_0") List<UnsignedByte> byteQ0Args,
    @Singular @JsonProperty("BYTE_Q_1") List<UnsignedByte> byteQ1Args,
    @Singular @JsonProperty("BYTE_Q_2") List<UnsignedByte> byteQ2Args,
    @Singular @JsonProperty("BYTE_Q_3") List<UnsignedByte> byteQ3Args,
    @Singular @JsonProperty("BYTE_R_0") List<UnsignedByte> byteR0Args,
    @Singular @JsonProperty("BYTE_R_1") List<UnsignedByte> byteR1Args,
    @Singular @JsonProperty("BYTE_R_2") List<UnsignedByte> byteR2Args,
    @Singular @JsonProperty("BYTE_R_3") List<UnsignedByte> byteR3Args,
    @Singular @JsonProperty("CMP_1") List<Boolean> cmp1Args,
    @Singular @JsonProperty("CMP_2") List<Boolean> cmp2Args,
    @Singular @JsonProperty("CT") List<Integer> ctArgs,
    @Singular @JsonProperty("DEC_OUTPUT") List<Boolean> decOutputArgs,
    @Singular @JsonProperty("DEC_SIGNED") List<Boolean> decSignedArgs,
    @Singular @JsonProperty("INST") List<UnsignedByte> instArgs,
    @Singular @JsonProperty("MSB_1") List<Boolean> msb1Args,
    @Singular @JsonProperty("MSB_2") List<Boolean> msb2Args,
    @Singular @JsonProperty("OLI") List<Boolean> oliArgs,
    @Singular @JsonProperty("RES_HI") List<BigInteger> resHiArgs,
    @Singular @JsonProperty("RES_LO") List<BigInteger> resLoArgs,
    @Singular @JsonProperty("STAMP") List<Integer> modStampArgs) {}
