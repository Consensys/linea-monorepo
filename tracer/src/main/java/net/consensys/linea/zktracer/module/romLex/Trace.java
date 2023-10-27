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

package net.consensys.linea.zktracer.module.romLex;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public record Trace(
    @JsonProperty("ADDR_HI") List<BigInteger> addrHi,
    @JsonProperty("ADDR_LO") List<BigInteger> addrLo,
    @JsonProperty("CODE_FRAGMENT_INDEX") List<BigInteger> codeFragmentIndex,
    @JsonProperty("CODE_FRAGMENT_INDEX_INFTY") List<BigInteger> codeFragmentIndexInfty,
    @JsonProperty("CODE_SIZE") List<BigInteger> codeSize,
    @JsonProperty("COMMIT_TO_STATE") List<Boolean> commitToState,
    @JsonProperty("DEP_NUMBER") List<BigInteger> depNumber,
    @JsonProperty("DEP_STATUS") List<Boolean> depStatus,
    @JsonProperty("READ_FROM_STATE") List<Boolean> readFromState) {
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.addrHi.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ADDR_HI")
    private final List<BigInteger> addrHi;

    @JsonProperty("ADDR_LO")
    private final List<BigInteger> addrLo;

    @JsonProperty("CODE_FRAGMENT_INDEX")
    private final List<BigInteger> codeFragmentIndex;

    @JsonProperty("CODE_FRAGMENT_INDEX_INFTY")
    private final List<BigInteger> codeFragmentIndexInfty;

    @JsonProperty("CODE_SIZE")
    private final List<BigInteger> codeSize;

    @JsonProperty("COMMIT_TO_STATE")
    private final List<Boolean> commitToState;

    @JsonProperty("DEP_NUMBER")
    private final List<BigInteger> depNumber;

    @JsonProperty("DEP_STATUS")
    private final List<Boolean> depStatus;

    @JsonProperty("READ_FROM_STATE")
    private final List<Boolean> readFromState;

    private TraceBuilder(int length) {
      this.addrHi = new ArrayList<>(length);
      this.addrLo = new ArrayList<>(length);
      this.codeFragmentIndex = new ArrayList<>(length);
      this.codeFragmentIndexInfty = new ArrayList<>(length);
      this.codeSize = new ArrayList<>(length);
      this.commitToState = new ArrayList<>(length);
      this.depNumber = new ArrayList<>(length);
      this.depStatus = new ArrayList<>(length);
      this.readFromState = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.addrHi.size();
    }

    public TraceBuilder addrHi(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ADDR_HI already set");
      } else {
        filled.set(0);
      }

      addrHi.add(b);

      return this;
    }

    public TraceBuilder addrLo(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ADDR_LO already set");
      } else {
        filled.set(1);
      }

      addrLo.add(b);

      return this;
    }

    public TraceBuilder codeFragmentIndex(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX already set");
      } else {
        filled.set(2);
      }

      codeFragmentIndex.add(b);

      return this;
    }

    public TraceBuilder codeFragmentIndexInfty(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX_INFTY already set");
      } else {
        filled.set(3);
      }

      codeFragmentIndexInfty.add(b);

      return this;
    }

    public TraceBuilder codeSize(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("CODE_SIZE already set");
      } else {
        filled.set(4);
      }

      codeSize.add(b);

      return this;
    }

    public TraceBuilder commitToState(final Boolean b) {
      if (filled.get(5)) {
        throw new IllegalStateException("COMMIT_TO_STATE already set");
      } else {
        filled.set(5);
      }

      commitToState.add(b);

      return this;
    }

    public TraceBuilder depNumber(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("DEP_NUMBER already set");
      } else {
        filled.set(6);
      }

      depNumber.add(b);

      return this;
    }

    public TraceBuilder depStatus(final Boolean b) {
      if (filled.get(7)) {
        throw new IllegalStateException("DEP_STATUS already set");
      } else {
        filled.set(7);
      }

      depStatus.add(b);

      return this;
    }

    public TraceBuilder readFromState(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("READ_FROM_STATE already set");
      } else {
        filled.set(8);
      }

      readFromState.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ADDR_HI has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ADDR_LO has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX_INFTY has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("CODE_SIZE has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("COMMIT_TO_STATE has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("DEP_NUMBER has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("DEP_STATUS has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("READ_FROM_STATE has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        addrHi.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        addrLo.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        codeFragmentIndex.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        codeFragmentIndexInfty.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        codeSize.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        commitToState.add(false);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        depNumber.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        depStatus.add(false);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        readFromState.add(false);
        this.filled.set(8);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          addrHi,
          addrLo,
          codeFragmentIndex,
          codeFragmentIndexInfty,
          codeSize,
          commitToState,
          depNumber,
          depStatus,
          readFromState);
    }
  }
}
