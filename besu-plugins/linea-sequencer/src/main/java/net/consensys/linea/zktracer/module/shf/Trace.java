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

package net.consensys.linea.zktracer.module.shf;

import java.math.BigInteger;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import lombok.Builder;
import lombok.Singular;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({
  "ACC_1",
  "ACC_2",
  "ACC_3",
  "ACC_4",
  "ACC_5",
  "ARG_1_HI",
  "ARG_1_LO",
  "ARG_2_HI",
  "ARG_2_LO",
  "BITS",
  "BIT_1",
  "BIT_2",
  "BIT_3",
  "BIT_4",
  "BIT_B_3",
  "BIT_B_4",
  "BIT_B_5",
  "BIT_B_6",
  "BIT_B_7",
  "BYTE_1",
  "BYTE_2",
  "BYTE_3",
  "BYTE_4",
  "BYTE_5",
  "COUNTER",
  "INST",
  "IS_DATA",
  "KNOWN",
  "LEFT_ALIGNED_SUFFIX_HIGH",
  "LEFT_ALIGNED_SUFFIX_LOW",
  "LOW_3",
  "MICRO_SHIFT_PARAMETER",
  "NEG",
  "ONES",
  "ONE_LINE_INSTRUCTION",
  "RES_HI",
  "RES_LO",
  "RIGHT_ALIGNED_PREFIX_HIGH",
  "RIGHT_ALIGNED_PREFIX_LOW",
  "SHB_3_HI",
  "SHB_3_LO",
  "SHB_4_HI",
  "SHB_4_LO",
  "SHB_5_HI",
  "SHB_5_LO",
  "SHB_6_HI",
  "SHB_6_LO",
  "SHB_7_HI",
  "SHB_7_LO",
  "SHIFT_DIRECTION",
  "SHIFT_STAMP",
})
@Builder
record Trace(
    @Singular @JsonProperty("ACC_1") List<BigInteger> acc1Args,
    @Singular @JsonProperty("ACC_2") List<BigInteger> acc2Args,
    @Singular @JsonProperty("ACC_3") List<BigInteger> acc3Args,
    @Singular @JsonProperty("ACC_4") List<BigInteger> acc4Args,
    @Singular @JsonProperty("ACC_5") List<BigInteger> acc5Args,
    @Singular @JsonProperty("ARG_1_HI") List<BigInteger> arg1HiArgs,
    @Singular @JsonProperty("ARG_1_LO") List<BigInteger> arg1LoArgs,
    @Singular @JsonProperty("ARG_2_HI") List<BigInteger> arg2HiArgs,
    @Singular @JsonProperty("ARG_2_LO") List<BigInteger> arg2LoArgs,
    @Singular @JsonProperty("BITS") List<Boolean> bitsArgs,
    @Singular @JsonProperty("BIT_1") List<Boolean> bit1Args,
    @Singular @JsonProperty("BIT_2") List<Boolean> bit2Args,
    @Singular @JsonProperty("BIT_3") List<Boolean> bit3Args,
    @Singular @JsonProperty("BIT_4") List<Boolean> bit4Args,
    @Singular @JsonProperty("BIT_B_3") List<Boolean> bitB3Args,
    @Singular @JsonProperty("BIT_B_4") List<Boolean> bitB4Args,
    @Singular @JsonProperty("BIT_B_5") List<Boolean> bitB5Args,
    @Singular @JsonProperty("BIT_B_6") List<Boolean> bitB6Args,
    @Singular @JsonProperty("BIT_B_7") List<Boolean> bitB7Args,
    @Singular @JsonProperty("BYTE_1") List<UnsignedByte> byte1Args,
    @Singular @JsonProperty("BYTE_2") List<UnsignedByte> byte2Args,
    @Singular @JsonProperty("BYTE_3") List<UnsignedByte> byte3Args,
    @Singular @JsonProperty("BYTE_4") List<UnsignedByte> byte4Args,
    @Singular @JsonProperty("BYTE_5") List<UnsignedByte> byte5Args,
    @Singular @JsonProperty("COUNTER") List<Integer> counterArgs,
    @Singular @JsonProperty("INST") List<UnsignedByte> instArgs,
    @Singular @JsonProperty("IS_DATA") List<Boolean> isDataArgs,
    @Singular @JsonProperty("KNOWN") List<Boolean> knownArgs,
    @Singular @JsonProperty("LEFT_ALIGNED_SUFFIX_HIGH")
        List<UnsignedByte> leftAlignedSuffixHighArgs,
    @Singular @JsonProperty("LEFT_ALIGNED_SUFFIX_LOW") List<UnsignedByte> leftAlignedSuffixLowArgs,
    @Singular @JsonProperty("LOW_3") List<UnsignedByte> low3Args,
    @Singular @JsonProperty("MICRO_SHIFT_PARAMETER") List<UnsignedByte> microShiftParameterArgs,
    @Singular @JsonProperty("NEG") List<Boolean> negArgs,
    @Singular @JsonProperty("ONES") List<UnsignedByte> onesArgs,
    @Singular @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> oneLineInstructionArgs,
    @Singular @JsonProperty("RES_HI") List<BigInteger> resHiArgs,
    @Singular @JsonProperty("RES_LO") List<BigInteger> resLoArgs,
    @Singular @JsonProperty("RIGHT_ALIGNED_PREFIX_HIGH")
        List<UnsignedByte> rightAlignedPrefixHighArgs,
    @Singular @JsonProperty("RIGHT_ALIGNED_PREFIX_LOW")
        List<UnsignedByte> rightAlignedPrefixLowArgs,
    @Singular @JsonProperty("SHB_3_HI") List<UnsignedByte> shb3HiArgs,
    @Singular @JsonProperty("SHB_3_LO") List<UnsignedByte> shb3LoArgs,
    @Singular @JsonProperty("SHB_4_HI") List<UnsignedByte> shb4HiArgs,
    @Singular @JsonProperty("SHB_4_LO") List<UnsignedByte> shb4LoArgs,
    @Singular @JsonProperty("SHB_5_HI") List<UnsignedByte> shb5HiArgs,
    @Singular @JsonProperty("SHB_5_LO") List<UnsignedByte> shb5LoArgs,
    @Singular @JsonProperty("SHB_6_HI") List<UnsignedByte> shb6HiArgs,
    @Singular @JsonProperty("SHB_6_LO") List<UnsignedByte> shb6LoArgs,
    @Singular @JsonProperty("SHB_7_HI") List<UnsignedByte> shb7HiArgs,
    @Singular @JsonProperty("SHB_7_LO") List<UnsignedByte> shb7LoArgs,
    @Singular @JsonProperty("SHIFT_DIRECTION") List<Boolean> shiftDirectionArgs,
    @Singular @JsonProperty("SHIFT_STAMP") List<Integer> shiftStampArgs) {}
