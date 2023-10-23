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

package net.consensys.linea.zktracer.module.rlpAddr;

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
public record Trace(
    @JsonProperty("ACC") List<BigInteger> acc,
    @JsonProperty("ACC_BYTESIZE") List<BigInteger> accBytesize,
    @JsonProperty("ADDR_HI") List<BigInteger> addrHi,
    @JsonProperty("ADDR_LO") List<BigInteger> addrLo,
    @JsonProperty("BIT1") List<Boolean> bit1,
    @JsonProperty("BIT_ACC") List<UnsignedByte> bitAcc,
    @JsonProperty("BYTE1") List<UnsignedByte> byte1,
    @JsonProperty("COUNTER") List<BigInteger> counter,
    @JsonProperty("DEP_ADDR_HI") List<BigInteger> depAddrHi,
    @JsonProperty("DEP_ADDR_LO") List<BigInteger> depAddrLo,
    @JsonProperty("INDEX") List<BigInteger> index,
    @JsonProperty("KEC_HI") List<BigInteger> kecHi,
    @JsonProperty("KEC_LO") List<BigInteger> kecLo,
    @JsonProperty("LC") List<Boolean> lc,
    @JsonProperty("LIMB") List<BigInteger> limb,
    @JsonProperty("nBYTES") List<BigInteger> nBytes,
    @JsonProperty("NONCE") List<BigInteger> nonce,
    @JsonProperty("POWER") List<BigInteger> power,
    @JsonProperty("RECIPE") List<BigInteger> recipe,
    @JsonProperty("RECIPE_1") List<Boolean> recipe1,
    @JsonProperty("RECIPE_2") List<Boolean> recipe2,
    @JsonProperty("SALT_HI") List<BigInteger> saltHi,
    @JsonProperty("SALT_LO") List<BigInteger> saltLo,
    @JsonProperty("STAMP") List<BigInteger> stamp,
    @JsonProperty("TINY_NON_ZERO_NONCE") List<Boolean> tinyNonZeroNonce) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  public int size() {
    return this.acc.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ACC")
    private final List<BigInteger> acc = new ArrayList<>();

    @JsonProperty("ACC_BYTESIZE")
    private final List<BigInteger> accBytesize = new ArrayList<>();

    @JsonProperty("ADDR_HI")
    private final List<BigInteger> addrHi = new ArrayList<>();

    @JsonProperty("ADDR_LO")
    private final List<BigInteger> addrLo = new ArrayList<>();

    @JsonProperty("BIT1")
    private final List<Boolean> bit1 = new ArrayList<>();

    @JsonProperty("BIT_ACC")
    private final List<UnsignedByte> bitAcc = new ArrayList<>();

    @JsonProperty("BYTE1")
    private final List<UnsignedByte> byte1 = new ArrayList<>();

    @JsonProperty("COUNTER")
    private final List<BigInteger> counter = new ArrayList<>();

    @JsonProperty("DEP_ADDR_HI")
    private final List<BigInteger> depAddrHi = new ArrayList<>();

    @JsonProperty("DEP_ADDR_LO")
    private final List<BigInteger> depAddrLo = new ArrayList<>();

    @JsonProperty("INDEX")
    private final List<BigInteger> index = new ArrayList<>();

    @JsonProperty("KEC_HI")
    private final List<BigInteger> kecHi = new ArrayList<>();

    @JsonProperty("KEC_LO")
    private final List<BigInteger> kecLo = new ArrayList<>();

    @JsonProperty("LC")
    private final List<Boolean> lc = new ArrayList<>();

    @JsonProperty("LIMB")
    private final List<BigInteger> limb = new ArrayList<>();

    @JsonProperty("nBYTES")
    private final List<BigInteger> nBytes = new ArrayList<>();

    @JsonProperty("NONCE")
    private final List<BigInteger> nonce = new ArrayList<>();

    @JsonProperty("POWER")
    private final List<BigInteger> power = new ArrayList<>();

    @JsonProperty("RECIPE")
    private final List<BigInteger> recipe = new ArrayList<>();

    @JsonProperty("RECIPE_1")
    private final List<Boolean> recipe1 = new ArrayList<>();

    @JsonProperty("RECIPE_2")
    private final List<Boolean> recipe2 = new ArrayList<>();

    @JsonProperty("SALT_HI")
    private final List<BigInteger> saltHi = new ArrayList<>();

    @JsonProperty("SALT_LO")
    private final List<BigInteger> saltLo = new ArrayList<>();

    @JsonProperty("STAMP")
    private final List<BigInteger> stamp = new ArrayList<>();

    @JsonProperty("TINY_NON_ZERO_NONCE")
    private final List<Boolean> tinyNonZeroNonce = new ArrayList<>();

    private TraceBuilder() {}

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.acc.size();
    }

    public TraceBuilder acc(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC already set");
      } else {
        filled.set(0);
      }

      acc.add(b);

      return this;
    }

    public TraceBuilder accBytesize(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_BYTESIZE already set");
      } else {
        filled.set(1);
      }

      accBytesize.add(b);

      return this;
    }

    public TraceBuilder addrHi(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ADDR_HI already set");
      } else {
        filled.set(2);
      }

      addrHi.add(b);

      return this;
    }

    public TraceBuilder addrLo(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ADDR_LO already set");
      } else {
        filled.set(3);
      }

      addrLo.add(b);

      return this;
    }

    public TraceBuilder bit1(final Boolean b) {
      if (filled.get(4)) {
        throw new IllegalStateException("BIT1 already set");
      } else {
        filled.set(4);
      }

      bit1.add(b);

      return this;
    }

    public TraceBuilder bitAcc(final UnsignedByte b) {
      if (filled.get(5)) {
        throw new IllegalStateException("BIT_ACC already set");
      } else {
        filled.set(5);
      }

      bitAcc.add(b);

      return this;
    }

    public TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(6)) {
        throw new IllegalStateException("BYTE1 already set");
      } else {
        filled.set(6);
      }

      byte1.add(b);

      return this;
    }

    public TraceBuilder counter(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(7);
      }

      counter.add(b);

      return this;
    }

    public TraceBuilder depAddrHi(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("DEP_ADDR_HI already set");
      } else {
        filled.set(8);
      }

      depAddrHi.add(b);

      return this;
    }

    public TraceBuilder depAddrLo(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("DEP_ADDR_LO already set");
      } else {
        filled.set(9);
      }

      depAddrLo.add(b);

      return this;
    }

    public TraceBuilder index(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("INDEX already set");
      } else {
        filled.set(10);
      }

      index.add(b);

      return this;
    }

    public TraceBuilder kecHi(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("KEC_HI already set");
      } else {
        filled.set(11);
      }

      kecHi.add(b);

      return this;
    }

    public TraceBuilder kecLo(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("KEC_LO already set");
      } else {
        filled.set(12);
      }

      kecLo.add(b);

      return this;
    }

    public TraceBuilder lc(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("LC already set");
      } else {
        filled.set(13);
      }

      lc.add(b);

      return this;
    }

    public TraceBuilder limb(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("LIMB already set");
      } else {
        filled.set(14);
      }

      limb.add(b);

      return this;
    }

    public TraceBuilder nBytes(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("nBYTES already set");
      } else {
        filled.set(24);
      }

      nBytes.add(b);

      return this;
    }

    public TraceBuilder nonce(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("NONCE already set");
      } else {
        filled.set(15);
      }

      nonce.add(b);

      return this;
    }

    public TraceBuilder power(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("POWER already set");
      } else {
        filled.set(16);
      }

      power.add(b);

      return this;
    }

    public TraceBuilder recipe(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("RECIPE already set");
      } else {
        filled.set(17);
      }

      recipe.add(b);

      return this;
    }

    public TraceBuilder recipe1(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("RECIPE_1 already set");
      } else {
        filled.set(18);
      }

      recipe1.add(b);

      return this;
    }

    public TraceBuilder recipe2(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("RECIPE_2 already set");
      } else {
        filled.set(19);
      }

      recipe2.add(b);

      return this;
    }

    public TraceBuilder saltHi(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("SALT_HI already set");
      } else {
        filled.set(20);
      }

      saltHi.add(b);

      return this;
    }

    public TraceBuilder saltLo(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("SALT_LO already set");
      } else {
        filled.set(21);
      }

      saltLo.add(b);

      return this;
    }

    public TraceBuilder stamp(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(22);
      }

      stamp.add(b);

      return this;
    }

    public TraceBuilder tinyNonZeroNonce(final Boolean b) {
      if (filled.get(23)) {
        throw new IllegalStateException("TINY_NON_ZERO_NONCE already set");
      } else {
        filled.set(23);
      }

      tinyNonZeroNonce.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ACC has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_BYTESIZE has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ADDR_HI has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ADDR_LO has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("BIT1 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("BIT_ACC has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("BYTE1 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("DEP_ADDR_HI has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("DEP_ADDR_LO has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("INDEX has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("KEC_HI has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("KEC_LO has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("LC has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("LIMB has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("nBYTES has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("NONCE has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("POWER has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("RECIPE has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("RECIPE_1 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("RECIPE_2 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("SALT_HI has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("SALT_LO has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("STAMP has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("TINY_NON_ZERO_NONCE has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        acc.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        accBytesize.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        addrHi.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        addrLo.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        bit1.add(false);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        bitAcc.add(UnsignedByte.of(0));
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        counter.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        depAddrHi.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        depAddrLo.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        index.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        kecHi.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        kecLo.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        lc.add(false);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        limb.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(24)) {
        nBytes.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(15)) {
        nonce.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        power.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        recipe.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        recipe1.add(false);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        recipe2.add(false);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        saltHi.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        saltLo.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        stamp.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        tinyNonZeroNonce.add(false);
        this.filled.set(23);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          acc,
          accBytesize,
          addrHi,
          addrLo,
          bit1,
          bitAcc,
          byte1,
          counter,
          depAddrHi,
          depAddrLo,
          index,
          kecHi,
          kecLo,
          lc,
          limb,
          nBytes,
          nonce,
          power,
          recipe,
          recipe1,
          recipe2,
          saltHi,
          saltLo,
          stamp,
          tinyNonZeroNonce);
    }
  }
}
