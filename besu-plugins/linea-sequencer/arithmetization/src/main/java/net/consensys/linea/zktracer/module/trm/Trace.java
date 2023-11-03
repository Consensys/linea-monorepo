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

package net.consensys.linea.zktracer.module.trm;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import net.consensys.linea.zktracer.types.UnsignedByte;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public record Trace(
    @JsonProperty("ACC_HI") List<BigInteger> accHi,
    @JsonProperty("ACC_LO") List<BigInteger> accLo,
    @JsonProperty("ACC_T") List<BigInteger> accT,
    @JsonProperty("ADDR_HI") List<BigInteger> addrHi,
    @JsonProperty("ADDR_LO") List<BigInteger> addrLo,
    @JsonProperty("BYTE_HI") List<UnsignedByte> byteHi,
    @JsonProperty("BYTE_LO") List<UnsignedByte> byteLo,
    @JsonProperty("CT") List<BigInteger> ct,
    @JsonProperty("IS_PREC") List<Boolean> isPrec,
    @JsonProperty("ONE") List<Boolean> one,
    @JsonProperty("PBIT") List<Boolean> pbit,
    @JsonProperty("STAMP") List<BigInteger> stamp,
    @JsonProperty("TRM_ADDR_HI") List<BigInteger> trmAddrHi) {
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.accHi.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ACC_HI")
    private final List<BigInteger> accHi;

    @JsonProperty("ACC_LO")
    private final List<BigInteger> accLo;

    @JsonProperty("ACC_T")
    private final List<BigInteger> accT;

    @JsonProperty("ADDR_HI")
    private final List<BigInteger> addrHi;

    @JsonProperty("ADDR_LO")
    private final List<BigInteger> addrLo;

    @JsonProperty("BYTE_HI")
    private final List<UnsignedByte> byteHi;

    @JsonProperty("BYTE_LO")
    private final List<UnsignedByte> byteLo;

    @JsonProperty("CT")
    private final List<BigInteger> ct;

    @JsonProperty("IS_PREC")
    private final List<Boolean> isPrec;

    @JsonProperty("ONE")
    private final List<Boolean> one;

    @JsonProperty("PBIT")
    private final List<Boolean> pbit;

    @JsonProperty("STAMP")
    private final List<BigInteger> stamp;

    @JsonProperty("TRM_ADDR_HI")
    private final List<BigInteger> trmAddrHi;

    private TraceBuilder(int length) {
      this.accHi = new ArrayList<>(length);
      this.accLo = new ArrayList<>(length);
      this.accT = new ArrayList<>(length);
      this.addrHi = new ArrayList<>(length);
      this.addrLo = new ArrayList<>(length);
      this.byteHi = new ArrayList<>(length);
      this.byteLo = new ArrayList<>(length);
      this.ct = new ArrayList<>(length);
      this.isPrec = new ArrayList<>(length);
      this.one = new ArrayList<>(length);
      this.pbit = new ArrayList<>(length);
      this.stamp = new ArrayList<>(length);
      this.trmAddrHi = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.accHi.size();
    }

    public TraceBuilder accHi(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC_HI already set");
      } else {
        filled.set(0);
      }

      accHi.add(b);

      return this;
    }

    public TraceBuilder accLo(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_LO already set");
      } else {
        filled.set(1);
      }

      accLo.add(b);

      return this;
    }

    public TraceBuilder accT(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ACC_T already set");
      } else {
        filled.set(2);
      }

      accT.add(b);

      return this;
    }

    public TraceBuilder addrHi(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ADDR_HI already set");
      } else {
        filled.set(3);
      }

      addrHi.add(b);

      return this;
    }

    public TraceBuilder addrLo(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ADDR_LO already set");
      } else {
        filled.set(4);
      }

      addrLo.add(b);

      return this;
    }

    public TraceBuilder byteHi(final UnsignedByte b) {
      if (filled.get(5)) {
        throw new IllegalStateException("BYTE_HI already set");
      } else {
        filled.set(5);
      }

      byteHi.add(b);

      return this;
    }

    public TraceBuilder byteLo(final UnsignedByte b) {
      if (filled.get(6)) {
        throw new IllegalStateException("BYTE_LO already set");
      } else {
        filled.set(6);
      }

      byteLo.add(b);

      return this;
    }

    public TraceBuilder ct(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(7);
      }

      ct.add(b);

      return this;
    }

    public TraceBuilder isPrec(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("IS_PREC already set");
      } else {
        filled.set(8);
      }

      isPrec.add(b);

      return this;
    }

    public TraceBuilder one(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ONE already set");
      } else {
        filled.set(9);
      }

      one.add(b);

      return this;
    }

    public TraceBuilder pbit(final Boolean b) {
      if (filled.get(10)) {
        throw new IllegalStateException("PBIT already set");
      } else {
        filled.set(10);
      }

      pbit.add(b);

      return this;
    }

    public TraceBuilder stamp(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(11);
      }

      stamp.add(b);

      return this;
    }

    public TraceBuilder trmAddrHi(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("TRM_ADDR_HI already set");
      } else {
        filled.set(12);
      }

      trmAddrHi.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ACC_HI has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_LO has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ACC_T has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ADDR_HI has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ADDR_LO has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("BYTE_HI has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("BYTE_LO has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("IS_PREC has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ONE has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("PBIT has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("STAMP has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("TRM_ADDR_HI has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        accHi.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        accLo.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        accT.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        addrHi.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        addrLo.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        byteHi.add(UnsignedByte.of(0));
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        byteLo.add(UnsignedByte.of(0));
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        ct.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        isPrec.add(false);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        one.add(false);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        pbit.add(false);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        stamp.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        trmAddrHi.add(BigInteger.ZERO);
        this.filled.set(12);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          accHi, accLo, accT, addrHi, addrLo, byteHi, byteLo, ct, isPrec, one, pbit, stamp,
          trmAddrHi);
    }
  }
}
