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
package net.consensys.linea.zktracer.module.module.alu.add;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import net.consensys.linea.zktracer.bytes.UnsignedByte;

@JsonPropertyOrder({"Trace", "Stamp"})
@SuppressWarnings("unused")
public record AddTrace (@JsonProperty("Trace") AddTrace.Trace trace, @JsonProperty("Stamp") int stamp) {
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
    @SuppressWarnings("unused")
    public record Trace(
            @JsonProperty("ACC_1") List<BigInteger> ACC_1,
            @JsonProperty("ACC_2") List<BigInteger> ACC_2,
            @JsonProperty("ARG_1_HI") List<BigInteger> ARG_1_HI,
            @JsonProperty("ARG_1_LO") List<BigInteger> ARG_1_LO,
            @JsonProperty("ARG_2_HI") List<BigInteger> ARG_2_HI,
            @JsonProperty("ARG_2_LO") List<BigInteger> ARG_2_LO,
            @JsonProperty("BYTE_1") List<UnsignedByte> BYTE_1,
            @JsonProperty("BYTE_2") List<UnsignedByte> BYTE_2,
            @JsonProperty("CT") List<Integer> COUNTER,
            @JsonProperty("INST") List<UnsignedByte> INST,
            @JsonProperty("OVERFLOW") List<Boolean> OVERFLOW,
            @JsonProperty("RES_HI") List<BigInteger> RES_HI,
            @JsonProperty("RES_LO") List<BigInteger> RES_LO,
            @JsonProperty("ADD_STAMP") List<Integer> ADD_STAMP) {

        public static class Builder {
            private final List<BigInteger> acc1 = new ArrayList<>();
            private final List<BigInteger> acc2 = new ArrayList<>();
            private final List<BigInteger> arg1Hi = new ArrayList<>();
            private final List<BigInteger> arg1Lo = new ArrayList<>();
            private final List<BigInteger> arg2Hi = new ArrayList<>();
            private final List<BigInteger> arg2Lo = new ArrayList<>();
            private final List<UnsignedByte> byte1 = new ArrayList<>();
            private final List<UnsignedByte> byte2 = new ArrayList<>();
            private final List<Integer> counter = new ArrayList<>();
            private final List<UnsignedByte> inst = new ArrayList<>();
            private final List<Boolean> overflow = new ArrayList<>();
            private final List<BigInteger> resHi = new ArrayList<>();
            private final List<BigInteger> resLo = new ArrayList<>();
            private final List<Integer> addStamp = new ArrayList<>();
            private int stamp = 0;

            private Builder() {
            }

            public static AddTrace.Trace.Builder newInstance() {
                return new AddTrace.Trace.Builder();
            }

            public AddTrace.Trace.Builder appendAcc1(final BigInteger b) {
                acc1.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendAcc2(final BigInteger b) {
                acc2.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendArg1Hi(final BigInteger b) {
                arg1Hi.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendArg1Lo(final BigInteger b) {
                arg1Lo.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendArg2Hi(final BigInteger b) {
                arg2Hi.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendArg2Lo(final BigInteger b) {
                arg2Lo.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendByte1(final UnsignedByte b) {
                byte1.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendByte2(final UnsignedByte b) {
                byte2.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendCounter(final Integer b) {
                counter.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendInst(final UnsignedByte b) {
                inst.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendOverflow(final Boolean b) {
                overflow.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendResHi(final BigInteger b) {
                resHi.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendResLo(final BigInteger b) {
                resLo.add(b);
                return this;
            }

            public AddTrace.Trace.Builder appendStamp(final Integer b) {
                addStamp.add(b);
                return this;
            }

            public AddTrace.Trace.Builder setStamp(final int stamp) {
                this.stamp = stamp;
                return this;
            }

            public AddTrace build() {
                return new AddTrace(
                        new AddTrace.Trace(
                                acc1,
                                acc2,
                                arg1Hi,
                                arg1Lo,
                                arg2Hi,
                                arg2Lo,
                                byte1,
                                byte2,
                                counter,
                                inst,
                                overflow,
                                resHi,
                                resLo,
                                addStamp),
                        stamp);

            }
        }
    }
}
