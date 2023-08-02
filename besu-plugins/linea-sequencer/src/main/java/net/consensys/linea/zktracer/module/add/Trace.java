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

package net.consensys.linea.zktracer.module.add;

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
record Trace(
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ARG_1_HI") List<BigInteger> arg1Hi,
    @JsonProperty("ARG_1_LO") List<BigInteger> arg1Lo,
    @JsonProperty("ARG_2_HI") List<BigInteger> arg2Hi,
    @JsonProperty("ARG_2_LO") List<BigInteger> arg2Lo,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("CT") List<BigInteger> ct,
    @JsonProperty("INST") List<BigInteger> inst,
    @JsonProperty("OVERFLOW") List<Boolean> overflow,
    @JsonProperty("RES_HI") List<BigInteger> resHi,
    @JsonProperty("RES_LO") List<BigInteger> resLo,
    @JsonProperty("STAMP") List<BigInteger> stamp) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    private final List<BigInteger> acc1 = new ArrayList<>();
    private final List<BigInteger> acc2 = new ArrayList<>();
    private final List<BigInteger> arg1Hi = new ArrayList<>();
    private final List<BigInteger> arg1Lo = new ArrayList<>();
    private final List<BigInteger> arg2Hi = new ArrayList<>();
    private final List<BigInteger> arg2Lo = new ArrayList<>();
    private final List<UnsignedByte> byte1 = new ArrayList<>();
    private final List<UnsignedByte> byte2 = new ArrayList<>();
    private final List<BigInteger> ct = new ArrayList<>();
    private final List<BigInteger> inst = new ArrayList<>();
    private final List<Boolean> overflow = new ArrayList<>();
    private final List<BigInteger> resHi = new ArrayList<>();
    private final List<BigInteger> resLo = new ArrayList<>();
    private final List<BigInteger> stamp = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder acc1(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(9);
      }

      acc1.add(b);

      return this;
    }

    TraceBuilder acc2(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(13);
      }

      acc2.add(b);

      return this;
    }

    TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(3);
      }

      arg1Hi.add(b);

      return this;
    }

    TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(11);
      }

      arg1Lo.add(b);

      return this;
    }

    TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(10);
      }

      arg2Hi.add(b);

      return this;
    }

    TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(0);
      }

      arg2Lo.add(b);

      return this;
    }

    TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(6)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(6);
      }

      byte1.add(b);

      return this;
    }

    TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(2)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(2);
      }

      byte2.add(b);

      return this;
    }

    TraceBuilder ct(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(4);
      }

      ct.add(b);

      return this;
    }

    TraceBuilder inst(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(7);
      }

      inst.add(b);

      return this;
    }

    TraceBuilder overflow(final Boolean b) {
      if (filled.get(1)) {
        throw new IllegalStateException("OVERFLOW already set");
      } else {
        filled.set(1);
      }

      overflow.add(b);

      return this;
    }

    TraceBuilder resHi(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(8);
      }

      resHi.add(b);

      return this;
    }

    TraceBuilder resLo(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(5);
      }

      resLo.add(b);

      return this;
    }

    TraceBuilder stamp(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(12);
      }

      stamp.add(b);

      return this;
    }

    TraceBuilder setAcc1At(final BigInteger b, int i) {
      acc1.set(i, b);

      return this;
    }

    TraceBuilder setAcc2At(final BigInteger b, int i) {
      acc2.set(i, b);

      return this;
    }

    TraceBuilder setArg1HiAt(final BigInteger b, int i) {
      arg1Hi.set(i, b);

      return this;
    }

    TraceBuilder setArg1LoAt(final BigInteger b, int i) {
      arg1Lo.set(i, b);

      return this;
    }

    TraceBuilder setArg2HiAt(final BigInteger b, int i) {
      arg2Hi.set(i, b);

      return this;
    }

    TraceBuilder setArg2LoAt(final BigInteger b, int i) {
      arg2Lo.set(i, b);

      return this;
    }

    TraceBuilder setByte1At(final UnsignedByte b, int i) {
      byte1.set(i, b);

      return this;
    }

    TraceBuilder setByte2At(final UnsignedByte b, int i) {
      byte2.set(i, b);

      return this;
    }

    TraceBuilder setCtAt(final BigInteger b, int i) {
      ct.set(i, b);

      return this;
    }

    TraceBuilder setInstAt(final BigInteger b, int i) {
      inst.set(i, b);

      return this;
    }

    TraceBuilder setOverflowAt(final Boolean b, int i) {
      overflow.set(i, b);

      return this;
    }

    TraceBuilder setResHiAt(final BigInteger b, int i) {
      resHi.set(i, b);

      return this;
    }

    TraceBuilder setResLoAt(final BigInteger b, int i) {
      resLo.set(i, b);

      return this;
    }

    TraceBuilder setStampAt(final BigInteger b, int i) {
      stamp.set(i, b);

      return this;
    }

    TraceBuilder setAcc1Relative(final BigInteger b, int i) {
      acc1.set(acc1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc2Relative(final BigInteger b, int i) {
      acc2.set(acc2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg1HiRelative(final BigInteger b, int i) {
      arg1Hi.set(arg1Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg1LoRelative(final BigInteger b, int i) {
      arg1Lo.set(arg1Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg2HiRelative(final BigInteger b, int i) {
      arg2Hi.set(arg2Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArg2LoRelative(final BigInteger b, int i) {
      arg2Lo.set(arg2Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte1Relative(final UnsignedByte b, int i) {
      byte1.set(byte1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte2Relative(final UnsignedByte b, int i) {
      byte2.set(byte2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCtRelative(final BigInteger b, int i) {
      ct.set(ct.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInstRelative(final BigInteger b, int i) {
      inst.set(inst.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOverflowRelative(final Boolean b, int i) {
      overflow.set(overflow.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResHiRelative(final BigInteger b, int i) {
      resHi.set(resHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResLoRelative(final BigInteger b, int i) {
      resLo.set(resLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setStampRelative(final BigInteger b, int i) {
      stamp.set(stamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(9)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("OVERFLOW has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("STAMP has not been filled");
      }

      filled.clear();

      return this;
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          acc1, acc2, arg1Hi, arg1Lo, arg2Hi, arg2Lo, byte1, byte2, ct, inst, overflow, resHi,
          resLo, stamp);
    }
  }
}
