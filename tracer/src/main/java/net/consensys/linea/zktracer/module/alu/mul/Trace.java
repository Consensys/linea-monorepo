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
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import lombok.Builder;
import lombok.Singular;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

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
@Builder
record Trace(
    @Singular @JsonProperty("ACC_A_0") List<BigInteger> accA0Args,
    @Singular @JsonProperty("ACC_A_1") List<BigInteger> accA1Args,
    @Singular @JsonProperty("ACC_A_2") List<BigInteger> accA2Args,
    @Singular @JsonProperty("ACC_A_3") List<BigInteger> accA3Args,
    @Singular @JsonProperty("ACC_B_0") List<BigInteger> accB0Args,
    @Singular @JsonProperty("ACC_B_1") List<BigInteger> accB1Args,
    @Singular @JsonProperty("ACC_B_2") List<BigInteger> accB2Args,
    @Singular @JsonProperty("ACC_B_3") List<BigInteger> accB3Args,
    @Singular @JsonProperty("ACC_C_0") List<BigInteger> accC0Args,
    @Singular @JsonProperty("ACC_C_1") List<BigInteger> accC1Args,
    @Singular @JsonProperty("ACC_C_2") List<BigInteger> accC2Args,
    @Singular @JsonProperty("ACC_C_3") List<BigInteger> accC3Args,
    @Singular @JsonProperty("ACC_H_0") List<BigInteger> accH0Args,
    @Singular @JsonProperty("ACC_H_1") List<BigInteger> accH1Args,
    @Singular @JsonProperty("ACC_H_2") List<BigInteger> accH2Args,
    @Singular @JsonProperty("ACC_H_3") List<BigInteger> accH3Args,
    @Singular @JsonProperty("ARG_1_HI") List<BigInteger> arg1HiArgs,
    @Singular @JsonProperty("ARG_1_LO") List<BigInteger> arg1LoArgs,
    @Singular @JsonProperty("ARG_2_HI") List<BigInteger> arg2HiArgs,
    @Singular @JsonProperty("ARG_2_LO") List<BigInteger> arg2LoArgs,
    @Singular @JsonProperty("BITS") List<Boolean> bitsArgs,
    @Singular @JsonProperty("BIT_NUM") List<Integer> bitNumArgs,
    @Singular @JsonProperty("BYTE_A_0") List<UnsignedByte> byteA0Args,
    @Singular @JsonProperty("BYTE_A_1") List<UnsignedByte> byteA1Args,
    @Singular @JsonProperty("BYTE_A_2") List<UnsignedByte> byteA2Args,
    @Singular @JsonProperty("BYTE_A_3") List<UnsignedByte> byteA3Args,
    @Singular @JsonProperty("BYTE_B_0") List<UnsignedByte> byteB0Args,
    @Singular @JsonProperty("BYTE_B_1") List<UnsignedByte> byteB1Args,
    @Singular @JsonProperty("BYTE_B_2") List<UnsignedByte> byteB2Args,
    @Singular @JsonProperty("BYTE_B_3") List<UnsignedByte> byteB3Args,
    @Singular @JsonProperty("BYTE_C_0") List<UnsignedByte> byteC0Args,
    @Singular @JsonProperty("BYTE_C_1") List<UnsignedByte> byteC1Args,
    @Singular @JsonProperty("BYTE_C_2") List<UnsignedByte> byteC2Args,
    @Singular @JsonProperty("BYTE_C_3") List<UnsignedByte> byteC3Args,
    @Singular @JsonProperty("BYTE_H_0") List<UnsignedByte> byteH0Args,
    @Singular @JsonProperty("BYTE_H_1") List<UnsignedByte> byteH1Args,
    @Singular @JsonProperty("BYTE_H_2") List<UnsignedByte> byteH2Args,
    @Singular @JsonProperty("BYTE_H_3") List<UnsignedByte> byteH3Args,
    @Singular @JsonProperty("COUNTER") List<Integer> counterArgs,
    @Singular @JsonProperty("EXPONENT_BIT") List<Boolean> exponentBitArgs,
    @Singular @JsonProperty("EXPONENT_BIT_ACCUMULATOR") List<BigInteger> exponentBitAccumulatorArgs,
    @Singular @JsonProperty("EXPONENT_BIT_SOURCE") List<Boolean> exponentBitSourceArgs,
    @Singular @JsonProperty("INST") List<UnsignedByte> instArgs,
    @Singular @JsonProperty("MUL_STAMP") List<Integer> mulStampArgs,
    @Singular @JsonProperty("OLI") List<Boolean> oneLineInstructionArgs,
    @Singular @JsonProperty("RESULT_VANISHES") List<Boolean> resultVanishesArgs,
    @Singular @JsonProperty("RES_HI") List<BigInteger> resHiArgs,
    @Singular @JsonProperty("RES_LO") List<BigInteger> resLoArgs,
    @Singular @JsonProperty("SQUARE_AND_MULTIPLY") List<Boolean> squareAndMultiplyArgs,
    @Singular @JsonProperty("TINY_BASE") List<Boolean> tinyBaseArgs,
    @Singular @JsonProperty("TINY_EXPONENT") List<Boolean> tinyExponentArgs) {}
