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

package net.consensys.linea.zktracer.module.tables.shf;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record Trace(
    @JsonProperty("BYTE") List<BigInteger> byteField,
    @JsonProperty("IS_IN_RT") List<BigInteger> isInRt,
    @JsonProperty("LAS") List<BigInteger> las,
    @JsonProperty("MSHP") List<BigInteger> mshp,
    @JsonProperty("ONES") List<BigInteger> ones,
    @JsonProperty("RAP") List<BigInteger> rap) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    private final List<BigInteger> byteField = new ArrayList<>();
    private final List<BigInteger> isInRt = new ArrayList<>();
    private final List<BigInteger> las = new ArrayList<>();
    private final List<BigInteger> mshp = new ArrayList<>();
    private final List<BigInteger> ones = new ArrayList<>();
    private final List<BigInteger> rap = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder byteField(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("BYTE already set");
      } else {
        filled.set(0);
      }

      byteField.add(b);

      return this;
    }

    TraceBuilder isInRt(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("IS_IN_RT already set");
      } else {
        filled.set(5);
      }

      isInRt.add(b);

      return this;
    }

    TraceBuilder las(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("LAS already set");
      } else {
        filled.set(1);
      }

      las.add(b);

      return this;
    }

    TraceBuilder mshp(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("MSHP already set");
      } else {
        filled.set(4);
      }

      mshp.add(b);

      return this;
    }

    TraceBuilder ones(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ONES already set");
      } else {
        filled.set(3);
      }

      ones.add(b);

      return this;
    }

    TraceBuilder rap(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("RAP already set");
      } else {
        filled.set(2);
      }

      rap.add(b);

      return this;
    }

    TraceBuilder setByteAt(final BigInteger b, int i) {
      byteField.set(i, b);

      return this;
    }

    TraceBuilder setIsInRtAt(final BigInteger b, int i) {
      isInRt.set(i, b);

      return this;
    }

    TraceBuilder setLasAt(final BigInteger b, int i) {
      las.set(i, b);

      return this;
    }

    TraceBuilder setMshpAt(final BigInteger b, int i) {
      mshp.set(i, b);

      return this;
    }

    TraceBuilder setOnesAt(final BigInteger b, int i) {
      ones.set(i, b);

      return this;
    }

    TraceBuilder setRapAt(final BigInteger b, int i) {
      rap.set(i, b);

      return this;
    }

    TraceBuilder setByteRelative(final BigInteger b, int i) {
      byteField.set(byteField.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsInRtRelative(final BigInteger b, int i) {
      isInRt.set(isInRt.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLasRelative(final BigInteger b, int i) {
      las.set(las.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setMshpRelative(final BigInteger b, int i) {
      mshp.set(mshp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOnesRelative(final BigInteger b, int i) {
      ones.set(ones.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setRapRelative(final BigInteger b, int i) {
      rap.set(rap.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("BYTE has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("IS_IN_RT has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("LAS has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("MSHP has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ONES has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("RAP has not been filled");
      }

      filled.clear();

      return this;
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(byteField, isInRt, las, mshp, ones, rap);
    }
  }
}
