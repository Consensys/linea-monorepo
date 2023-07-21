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

package net.consensys.linea.zktracer.module.wcp;

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
  "ACC_6",
  "ARGUMENT_1_HI",
  "ARGUMENT_1_LO",
  "ARGUMENT_2_HI",
  "ARGUMENT_2_LO",
  "BITS",
  "BIT_1",
  "BIT_2",
  "BIT_3",
  "BIT_4",
  "BYTE_1",
  "BYTE_2",
  "BYTE_3",
  "BYTE_4",
  "BYTE_5",
  "BYTE_6",
  "COUNTER",
  "INST",
  "NEG_1",
  "NEG_2",
  "ONE_LINE_INSTRUCTION",
  "RESULT_HI",
  "RESULT_LO",
  "WORD_COMPARISON_STAMP"
})
@Builder
record Trace(
    @Singular @JsonProperty("ACC_1") List<BigInteger> acc1Args,
    @Singular @JsonProperty("ACC_2") List<BigInteger> acc2Args,
    @Singular @JsonProperty("ACC_3") List<BigInteger> acc3Args,
    @Singular @JsonProperty("ACC_4") List<BigInteger> acc4Args,
    @Singular @JsonProperty("ACC_5") List<BigInteger> acc5Args,
    @Singular @JsonProperty("ACC_6") List<BigInteger> acc6Args,
    @Singular @JsonProperty("ARGUMENT_1_HI") List<BigInteger> argument1HiArgs,
    @Singular @JsonProperty("ARGUMENT_1_LO") List<BigInteger> argument1LoArgs,
    @Singular @JsonProperty("ARGUMENT_2_HI") List<BigInteger> argument2HiArgs,
    @Singular @JsonProperty("ARGUMENT_2_LO") List<BigInteger> argument2LoArgs,
    @Singular @JsonProperty("BITS") List<Boolean> bitsArgs,
    @Singular @JsonProperty("BIT_1") List<Boolean> bit1Args,
    @Singular @JsonProperty("BIT_2") List<Boolean> bit2Args,
    @Singular @JsonProperty("BIT_3") List<Boolean> bit3Args,
    @Singular @JsonProperty("BIT_4") List<Boolean> bit4Args,
    @Singular @JsonProperty("BYTE_1") List<UnsignedByte> byte1Args,
    @Singular @JsonProperty("BYTE_2") List<UnsignedByte> byte2Args,
    @Singular @JsonProperty("BYTE_3") List<UnsignedByte> byte3Args,
    @Singular @JsonProperty("BYTE_4") List<UnsignedByte> byte4Args,
    @Singular @JsonProperty("BYTE_5") List<UnsignedByte> byte5Args,
    @Singular @JsonProperty("BYTE_6") List<UnsignedByte> byte6Args,
    @Singular @JsonProperty("COUNTER") List<Integer> counterArgs,
    @Singular @JsonProperty("INST") List<UnsignedByte> instArgs,
    @Singular @JsonProperty("NEG_1") List<Boolean> neg1Args,
    @Singular @JsonProperty("NEG_2") List<Boolean> neg2Args,
    @Singular @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> oneLineInstructionArgs,
    @Singular @JsonProperty("RESULT_HI") List<Boolean> resultHiArgs,
    @Singular @JsonProperty("RESULT_LO") List<Boolean> resultLoArgs,
    @Singular @JsonProperty("WORD_COMPARISON_STAMP") List<Integer> wordComparisonStampArgs) {}
