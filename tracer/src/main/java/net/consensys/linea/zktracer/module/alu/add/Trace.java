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

package net.consensys.linea.zktracer.module.alu.add;

import java.math.BigInteger;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import lombok.Builder;
import lombok.Singular;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

/** Represents an execution trace for addition/subtraction operation. */
@JsonPropertyOrder({
  "ACC_1",
  "ACC_2",
  "ARG_1_HI",
  "ARG_1_LO",
  "ARG_2_HI",
  "ARG_2_LO",
  "BYTE_1",
  "BYTE_2",
  "CT", // IS CT SAME AS COUNTER?
  "INST",
  "OVERFLOW",
  "RES_HI",
  "RES_LO",
  "ADD_STAMP"
})
@Builder
record Trace(
    @Singular @JsonProperty("ACC_1") List<BigInteger> acc1Args,
    @Singular @JsonProperty("ACC_2") List<BigInteger> acc2Args,
    @Singular @JsonProperty("ARG_1_HI") List<BigInteger> arg1HiArgs,
    @Singular @JsonProperty("ARG_1_LO") List<BigInteger> arg1LoArgs,
    @Singular @JsonProperty("ARG_2_HI") List<BigInteger> arg2HiArgs,
    @Singular @JsonProperty("ARG_2_LO") List<BigInteger> arg2LoArgs,
    @Singular @JsonProperty("BYTE_1") List<UnsignedByte> byte1Args,
    @Singular @JsonProperty("BYTE_2") List<UnsignedByte> byte2Args,
    @Singular @JsonProperty("CT") List<Integer> counterArgs,
    @Singular @JsonProperty("INST") List<UnsignedByte> instArgs,
    @Singular @JsonProperty("OVERFLOW") List<Boolean> overflowArgs,
    @Singular @JsonProperty("RES_HI") List<BigInteger> resHiArgs,
    @Singular @JsonProperty("RES_LO") List<BigInteger> resLoArgs,
    @Singular @JsonProperty("ADD_STAMP") List<Integer> addStampArgs) {}
