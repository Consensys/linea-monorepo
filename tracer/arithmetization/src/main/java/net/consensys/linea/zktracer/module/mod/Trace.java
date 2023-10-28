/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.module.mod;

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
    @JsonProperty("ACC_1_2") List<BigInteger> acc12,
    @JsonProperty("ACC_1_3") List<BigInteger> acc13,
    @JsonProperty("ACC_2_2") List<BigInteger> acc22,
    @JsonProperty("ACC_2_3") List<BigInteger> acc23,
    @JsonProperty("ACC_B_0") List<BigInteger> accB0,
    @JsonProperty("ACC_B_1") List<BigInteger> accB1,
    @JsonProperty("ACC_B_2") List<BigInteger> accB2,
    @JsonProperty("ACC_B_3") List<BigInteger> accB3,
    @JsonProperty("ACC_DELTA_0") List<BigInteger> accDelta0,
    @JsonProperty("ACC_DELTA_1") List<BigInteger> accDelta1,
    @JsonProperty("ACC_DELTA_2") List<BigInteger> accDelta2,
    @JsonProperty("ACC_DELTA_3") List<BigInteger> accDelta3,
    @JsonProperty("ACC_H_0") List<BigInteger> accH0,
    @JsonProperty("ACC_H_1") List<BigInteger> accH1,
    @JsonProperty("ACC_H_2") List<BigInteger> accH2,
    @JsonProperty("ACC_Q_0") List<BigInteger> accQ0,
    @JsonProperty("ACC_Q_1") List<BigInteger> accQ1,
    @JsonProperty("ACC_Q_2") List<BigInteger> accQ2,
    @JsonProperty("ACC_Q_3") List<BigInteger> accQ3,
    @JsonProperty("ACC_R_0") List<BigInteger> accR0,
    @JsonProperty("ACC_R_1") List<BigInteger> accR1,
    @JsonProperty("ACC_R_2") List<BigInteger> accR2,
    @JsonProperty("ACC_R_3") List<BigInteger> accR3,
    @JsonProperty("ARG_1_HI") List<BigInteger> arg1Hi,
    @JsonProperty("ARG_1_LO") List<BigInteger> arg1Lo,
    @JsonProperty("ARG_2_HI") List<BigInteger> arg2Hi,
    @JsonProperty("ARG_2_LO") List<BigInteger> arg2Lo,
    @JsonProperty("BYTE_1_2") List<UnsignedByte> byte12,
    @JsonProperty("BYTE_1_3") List<UnsignedByte> byte13,
    @JsonProperty("BYTE_2_2") List<UnsignedByte> byte22,
    @JsonProperty("BYTE_2_3") List<UnsignedByte> byte23,
    @JsonProperty("BYTE_B_0") List<UnsignedByte> byteB0,
    @JsonProperty("BYTE_B_1") List<UnsignedByte> byteB1,
    @JsonProperty("BYTE_B_2") List<UnsignedByte> byteB2,
    @JsonProperty("BYTE_B_3") List<UnsignedByte> byteB3,
    @JsonProperty("BYTE_DELTA_0") List<UnsignedByte> byteDelta0,
    @JsonProperty("BYTE_DELTA_1") List<UnsignedByte> byteDelta1,
    @JsonProperty("BYTE_DELTA_2") List<UnsignedByte> byteDelta2,
    @JsonProperty("BYTE_DELTA_3") List<UnsignedByte> byteDelta3,
    @JsonProperty("BYTE_H_0") List<UnsignedByte> byteH0,
    @JsonProperty("BYTE_H_1") List<UnsignedByte> byteH1,
    @JsonProperty("BYTE_H_2") List<UnsignedByte> byteH2,
    @JsonProperty("BYTE_Q_0") List<UnsignedByte> byteQ0,
    @JsonProperty("BYTE_Q_1") List<UnsignedByte> byteQ1,
    @JsonProperty("BYTE_Q_2") List<UnsignedByte> byteQ2,
    @JsonProperty("BYTE_Q_3") List<UnsignedByte> byteQ3,
    @JsonProperty("BYTE_R_0") List<UnsignedByte> byteR0,
    @JsonProperty("BYTE_R_1") List<UnsignedByte> byteR1,
    @JsonProperty("BYTE_R_2") List<UnsignedByte> byteR2,
    @JsonProperty("BYTE_R_3") List<UnsignedByte> byteR3,
    @JsonProperty("CMP_1") List<Boolean> cmp1,
    @JsonProperty("CMP_2") List<Boolean> cmp2,
    @JsonProperty("CT") List<BigInteger> ct,
    @JsonProperty("DEC_OUTPUT") List<Boolean> decOutput,
    @JsonProperty("DEC_SIGNED") List<Boolean> decSigned,
    @JsonProperty("INST") List<BigInteger> inst,
    @JsonProperty("MSB_1") List<Boolean> msb1,
    @JsonProperty("MSB_2") List<Boolean> msb2,
    @JsonProperty("OLI") List<Boolean> oli,
    @JsonProperty("RES_HI") List<BigInteger> resHi,
    @JsonProperty("RES_LO") List<BigInteger> resLo,
    @JsonProperty("STAMP") List<BigInteger> stamp) {
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.acc12.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ACC_1_2")
    private final List<BigInteger> acc12;

    @JsonProperty("ACC_1_3")
    private final List<BigInteger> acc13;

    @JsonProperty("ACC_2_2")
    private final List<BigInteger> acc22;

    @JsonProperty("ACC_2_3")
    private final List<BigInteger> acc23;

    @JsonProperty("ACC_B_0")
    private final List<BigInteger> accB0;

    @JsonProperty("ACC_B_1")
    private final List<BigInteger> accB1;

    @JsonProperty("ACC_B_2")
    private final List<BigInteger> accB2;

    @JsonProperty("ACC_B_3")
    private final List<BigInteger> accB3;

    @JsonProperty("ACC_DELTA_0")
    private final List<BigInteger> accDelta0;

    @JsonProperty("ACC_DELTA_1")
    private final List<BigInteger> accDelta1;

    @JsonProperty("ACC_DELTA_2")
    private final List<BigInteger> accDelta2;

    @JsonProperty("ACC_DELTA_3")
    private final List<BigInteger> accDelta3;

    @JsonProperty("ACC_H_0")
    private final List<BigInteger> accH0;

    @JsonProperty("ACC_H_1")
    private final List<BigInteger> accH1;

    @JsonProperty("ACC_H_2")
    private final List<BigInteger> accH2;

    @JsonProperty("ACC_Q_0")
    private final List<BigInteger> accQ0;

    @JsonProperty("ACC_Q_1")
    private final List<BigInteger> accQ1;

    @JsonProperty("ACC_Q_2")
    private final List<BigInteger> accQ2;

    @JsonProperty("ACC_Q_3")
    private final List<BigInteger> accQ3;

    @JsonProperty("ACC_R_0")
    private final List<BigInteger> accR0;

    @JsonProperty("ACC_R_1")
    private final List<BigInteger> accR1;

    @JsonProperty("ACC_R_2")
    private final List<BigInteger> accR2;

    @JsonProperty("ACC_R_3")
    private final List<BigInteger> accR3;

    @JsonProperty("ARG_1_HI")
    private final List<BigInteger> arg1Hi;

    @JsonProperty("ARG_1_LO")
    private final List<BigInteger> arg1Lo;

    @JsonProperty("ARG_2_HI")
    private final List<BigInteger> arg2Hi;

    @JsonProperty("ARG_2_LO")
    private final List<BigInteger> arg2Lo;

    @JsonProperty("BYTE_1_2")
    private final List<UnsignedByte> byte12;

    @JsonProperty("BYTE_1_3")
    private final List<UnsignedByte> byte13;

    @JsonProperty("BYTE_2_2")
    private final List<UnsignedByte> byte22;

    @JsonProperty("BYTE_2_3")
    private final List<UnsignedByte> byte23;

    @JsonProperty("BYTE_B_0")
    private final List<UnsignedByte> byteB0;

    @JsonProperty("BYTE_B_1")
    private final List<UnsignedByte> byteB1;

    @JsonProperty("BYTE_B_2")
    private final List<UnsignedByte> byteB2;

    @JsonProperty("BYTE_B_3")
    private final List<UnsignedByte> byteB3;

    @JsonProperty("BYTE_DELTA_0")
    private final List<UnsignedByte> byteDelta0;

    @JsonProperty("BYTE_DELTA_1")
    private final List<UnsignedByte> byteDelta1;

    @JsonProperty("BYTE_DELTA_2")
    private final List<UnsignedByte> byteDelta2;

    @JsonProperty("BYTE_DELTA_3")
    private final List<UnsignedByte> byteDelta3;

    @JsonProperty("BYTE_H_0")
    private final List<UnsignedByte> byteH0;

    @JsonProperty("BYTE_H_1")
    private final List<UnsignedByte> byteH1;

    @JsonProperty("BYTE_H_2")
    private final List<UnsignedByte> byteH2;

    @JsonProperty("BYTE_Q_0")
    private final List<UnsignedByte> byteQ0;

    @JsonProperty("BYTE_Q_1")
    private final List<UnsignedByte> byteQ1;

    @JsonProperty("BYTE_Q_2")
    private final List<UnsignedByte> byteQ2;

    @JsonProperty("BYTE_Q_3")
    private final List<UnsignedByte> byteQ3;

    @JsonProperty("BYTE_R_0")
    private final List<UnsignedByte> byteR0;

    @JsonProperty("BYTE_R_1")
    private final List<UnsignedByte> byteR1;

    @JsonProperty("BYTE_R_2")
    private final List<UnsignedByte> byteR2;

    @JsonProperty("BYTE_R_3")
    private final List<UnsignedByte> byteR3;

    @JsonProperty("CMP_1")
    private final List<Boolean> cmp1;

    @JsonProperty("CMP_2")
    private final List<Boolean> cmp2;

    @JsonProperty("CT")
    private final List<BigInteger> ct;

    @JsonProperty("DEC_OUTPUT")
    private final List<Boolean> decOutput;

    @JsonProperty("DEC_SIGNED")
    private final List<Boolean> decSigned;

    @JsonProperty("INST")
    private final List<BigInteger> inst;

    @JsonProperty("MSB_1")
    private final List<Boolean> msb1;

    @JsonProperty("MSB_2")
    private final List<Boolean> msb2;

    @JsonProperty("OLI")
    private final List<Boolean> oli;

    @JsonProperty("RES_HI")
    private final List<BigInteger> resHi;

    @JsonProperty("RES_LO")
    private final List<BigInteger> resLo;

    @JsonProperty("STAMP")
    private final List<BigInteger> stamp;

    private TraceBuilder(int length) {
      this.acc12 = new ArrayList<>(length);
      this.acc13 = new ArrayList<>(length);
      this.acc22 = new ArrayList<>(length);
      this.acc23 = new ArrayList<>(length);
      this.accB0 = new ArrayList<>(length);
      this.accB1 = new ArrayList<>(length);
      this.accB2 = new ArrayList<>(length);
      this.accB3 = new ArrayList<>(length);
      this.accDelta0 = new ArrayList<>(length);
      this.accDelta1 = new ArrayList<>(length);
      this.accDelta2 = new ArrayList<>(length);
      this.accDelta3 = new ArrayList<>(length);
      this.accH0 = new ArrayList<>(length);
      this.accH1 = new ArrayList<>(length);
      this.accH2 = new ArrayList<>(length);
      this.accQ0 = new ArrayList<>(length);
      this.accQ1 = new ArrayList<>(length);
      this.accQ2 = new ArrayList<>(length);
      this.accQ3 = new ArrayList<>(length);
      this.accR0 = new ArrayList<>(length);
      this.accR1 = new ArrayList<>(length);
      this.accR2 = new ArrayList<>(length);
      this.accR3 = new ArrayList<>(length);
      this.arg1Hi = new ArrayList<>(length);
      this.arg1Lo = new ArrayList<>(length);
      this.arg2Hi = new ArrayList<>(length);
      this.arg2Lo = new ArrayList<>(length);
      this.byte12 = new ArrayList<>(length);
      this.byte13 = new ArrayList<>(length);
      this.byte22 = new ArrayList<>(length);
      this.byte23 = new ArrayList<>(length);
      this.byteB0 = new ArrayList<>(length);
      this.byteB1 = new ArrayList<>(length);
      this.byteB2 = new ArrayList<>(length);
      this.byteB3 = new ArrayList<>(length);
      this.byteDelta0 = new ArrayList<>(length);
      this.byteDelta1 = new ArrayList<>(length);
      this.byteDelta2 = new ArrayList<>(length);
      this.byteDelta3 = new ArrayList<>(length);
      this.byteH0 = new ArrayList<>(length);
      this.byteH1 = new ArrayList<>(length);
      this.byteH2 = new ArrayList<>(length);
      this.byteQ0 = new ArrayList<>(length);
      this.byteQ1 = new ArrayList<>(length);
      this.byteQ2 = new ArrayList<>(length);
      this.byteQ3 = new ArrayList<>(length);
      this.byteR0 = new ArrayList<>(length);
      this.byteR1 = new ArrayList<>(length);
      this.byteR2 = new ArrayList<>(length);
      this.byteR3 = new ArrayList<>(length);
      this.cmp1 = new ArrayList<>(length);
      this.cmp2 = new ArrayList<>(length);
      this.ct = new ArrayList<>(length);
      this.decOutput = new ArrayList<>(length);
      this.decSigned = new ArrayList<>(length);
      this.inst = new ArrayList<>(length);
      this.msb1 = new ArrayList<>(length);
      this.msb2 = new ArrayList<>(length);
      this.oli = new ArrayList<>(length);
      this.resHi = new ArrayList<>(length);
      this.resLo = new ArrayList<>(length);
      this.stamp = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.acc12.size();
    }

    public TraceBuilder acc12(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC_1_2 already set");
      } else {
        filled.set(0);
      }

      acc12.add(b);

      return this;
    }

    public TraceBuilder acc13(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_1_3 already set");
      } else {
        filled.set(1);
      }

      acc13.add(b);

      return this;
    }

    public TraceBuilder acc22(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ACC_2_2 already set");
      } else {
        filled.set(2);
      }

      acc22.add(b);

      return this;
    }

    public TraceBuilder acc23(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_2_3 already set");
      } else {
        filled.set(3);
      }

      acc23.add(b);

      return this;
    }

    public TraceBuilder accB0(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_B_0 already set");
      } else {
        filled.set(4);
      }

      accB0.add(b);

      return this;
    }

    public TraceBuilder accB1(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("ACC_B_1 already set");
      } else {
        filled.set(5);
      }

      accB1.add(b);

      return this;
    }

    public TraceBuilder accB2(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ACC_B_2 already set");
      } else {
        filled.set(6);
      }

      accB2.add(b);

      return this;
    }

    public TraceBuilder accB3(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ACC_B_3 already set");
      } else {
        filled.set(7);
      }

      accB3.add(b);

      return this;
    }

    public TraceBuilder accDelta0(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ACC_DELTA_0 already set");
      } else {
        filled.set(8);
      }

      accDelta0.add(b);

      return this;
    }

    public TraceBuilder accDelta1(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ACC_DELTA_1 already set");
      } else {
        filled.set(9);
      }

      accDelta1.add(b);

      return this;
    }

    public TraceBuilder accDelta2(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("ACC_DELTA_2 already set");
      } else {
        filled.set(10);
      }

      accDelta2.add(b);

      return this;
    }

    public TraceBuilder accDelta3(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("ACC_DELTA_3 already set");
      } else {
        filled.set(11);
      }

      accDelta3.add(b);

      return this;
    }

    public TraceBuilder accH0(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("ACC_H_0 already set");
      } else {
        filled.set(12);
      }

      accH0.add(b);

      return this;
    }

    public TraceBuilder accH1(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("ACC_H_1 already set");
      } else {
        filled.set(13);
      }

      accH1.add(b);

      return this;
    }

    public TraceBuilder accH2(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("ACC_H_2 already set");
      } else {
        filled.set(14);
      }

      accH2.add(b);

      return this;
    }

    public TraceBuilder accQ0(final BigInteger b) {
      if (filled.get(15)) {
        throw new IllegalStateException("ACC_Q_0 already set");
      } else {
        filled.set(15);
      }

      accQ0.add(b);

      return this;
    }

    public TraceBuilder accQ1(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("ACC_Q_1 already set");
      } else {
        filled.set(16);
      }

      accQ1.add(b);

      return this;
    }

    public TraceBuilder accQ2(final BigInteger b) {
      if (filled.get(17)) {
        throw new IllegalStateException("ACC_Q_2 already set");
      } else {
        filled.set(17);
      }

      accQ2.add(b);

      return this;
    }

    public TraceBuilder accQ3(final BigInteger b) {
      if (filled.get(18)) {
        throw new IllegalStateException("ACC_Q_3 already set");
      } else {
        filled.set(18);
      }

      accQ3.add(b);

      return this;
    }

    public TraceBuilder accR0(final BigInteger b) {
      if (filled.get(19)) {
        throw new IllegalStateException("ACC_R_0 already set");
      } else {
        filled.set(19);
      }

      accR0.add(b);

      return this;
    }

    public TraceBuilder accR1(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("ACC_R_1 already set");
      } else {
        filled.set(20);
      }

      accR1.add(b);

      return this;
    }

    public TraceBuilder accR2(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("ACC_R_2 already set");
      } else {
        filled.set(21);
      }

      accR2.add(b);

      return this;
    }

    public TraceBuilder accR3(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("ACC_R_3 already set");
      } else {
        filled.set(22);
      }

      accR3.add(b);

      return this;
    }

    public TraceBuilder arg1Hi(final BigInteger b) {
      if (filled.get(23)) {
        throw new IllegalStateException("ARG_1_HI already set");
      } else {
        filled.set(23);
      }

      arg1Hi.add(b);

      return this;
    }

    public TraceBuilder arg1Lo(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("ARG_1_LO already set");
      } else {
        filled.set(24);
      }

      arg1Lo.add(b);

      return this;
    }

    public TraceBuilder arg2Hi(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ARG_2_HI already set");
      } else {
        filled.set(25);
      }

      arg2Hi.add(b);

      return this;
    }

    public TraceBuilder arg2Lo(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("ARG_2_LO already set");
      } else {
        filled.set(26);
      }

      arg2Lo.add(b);

      return this;
    }

    public TraceBuilder byte12(final UnsignedByte b) {
      if (filled.get(27)) {
        throw new IllegalStateException("BYTE_1_2 already set");
      } else {
        filled.set(27);
      }

      byte12.add(b);

      return this;
    }

    public TraceBuilder byte13(final UnsignedByte b) {
      if (filled.get(28)) {
        throw new IllegalStateException("BYTE_1_3 already set");
      } else {
        filled.set(28);
      }

      byte13.add(b);

      return this;
    }

    public TraceBuilder byte22(final UnsignedByte b) {
      if (filled.get(29)) {
        throw new IllegalStateException("BYTE_2_2 already set");
      } else {
        filled.set(29);
      }

      byte22.add(b);

      return this;
    }

    public TraceBuilder byte23(final UnsignedByte b) {
      if (filled.get(30)) {
        throw new IllegalStateException("BYTE_2_3 already set");
      } else {
        filled.set(30);
      }

      byte23.add(b);

      return this;
    }

    public TraceBuilder byteB0(final UnsignedByte b) {
      if (filled.get(31)) {
        throw new IllegalStateException("BYTE_B_0 already set");
      } else {
        filled.set(31);
      }

      byteB0.add(b);

      return this;
    }

    public TraceBuilder byteB1(final UnsignedByte b) {
      if (filled.get(32)) {
        throw new IllegalStateException("BYTE_B_1 already set");
      } else {
        filled.set(32);
      }

      byteB1.add(b);

      return this;
    }

    public TraceBuilder byteB2(final UnsignedByte b) {
      if (filled.get(33)) {
        throw new IllegalStateException("BYTE_B_2 already set");
      } else {
        filled.set(33);
      }

      byteB2.add(b);

      return this;
    }

    public TraceBuilder byteB3(final UnsignedByte b) {
      if (filled.get(34)) {
        throw new IllegalStateException("BYTE_B_3 already set");
      } else {
        filled.set(34);
      }

      byteB3.add(b);

      return this;
    }

    public TraceBuilder byteDelta0(final UnsignedByte b) {
      if (filled.get(35)) {
        throw new IllegalStateException("BYTE_DELTA_0 already set");
      } else {
        filled.set(35);
      }

      byteDelta0.add(b);

      return this;
    }

    public TraceBuilder byteDelta1(final UnsignedByte b) {
      if (filled.get(36)) {
        throw new IllegalStateException("BYTE_DELTA_1 already set");
      } else {
        filled.set(36);
      }

      byteDelta1.add(b);

      return this;
    }

    public TraceBuilder byteDelta2(final UnsignedByte b) {
      if (filled.get(37)) {
        throw new IllegalStateException("BYTE_DELTA_2 already set");
      } else {
        filled.set(37);
      }

      byteDelta2.add(b);

      return this;
    }

    public TraceBuilder byteDelta3(final UnsignedByte b) {
      if (filled.get(38)) {
        throw new IllegalStateException("BYTE_DELTA_3 already set");
      } else {
        filled.set(38);
      }

      byteDelta3.add(b);

      return this;
    }

    public TraceBuilder byteH0(final UnsignedByte b) {
      if (filled.get(39)) {
        throw new IllegalStateException("BYTE_H_0 already set");
      } else {
        filled.set(39);
      }

      byteH0.add(b);

      return this;
    }

    public TraceBuilder byteH1(final UnsignedByte b) {
      if (filled.get(40)) {
        throw new IllegalStateException("BYTE_H_1 already set");
      } else {
        filled.set(40);
      }

      byteH1.add(b);

      return this;
    }

    public TraceBuilder byteH2(final UnsignedByte b) {
      if (filled.get(41)) {
        throw new IllegalStateException("BYTE_H_2 already set");
      } else {
        filled.set(41);
      }

      byteH2.add(b);

      return this;
    }

    public TraceBuilder byteQ0(final UnsignedByte b) {
      if (filled.get(42)) {
        throw new IllegalStateException("BYTE_Q_0 already set");
      } else {
        filled.set(42);
      }

      byteQ0.add(b);

      return this;
    }

    public TraceBuilder byteQ1(final UnsignedByte b) {
      if (filled.get(43)) {
        throw new IllegalStateException("BYTE_Q_1 already set");
      } else {
        filled.set(43);
      }

      byteQ1.add(b);

      return this;
    }

    public TraceBuilder byteQ2(final UnsignedByte b) {
      if (filled.get(44)) {
        throw new IllegalStateException("BYTE_Q_2 already set");
      } else {
        filled.set(44);
      }

      byteQ2.add(b);

      return this;
    }

    public TraceBuilder byteQ3(final UnsignedByte b) {
      if (filled.get(45)) {
        throw new IllegalStateException("BYTE_Q_3 already set");
      } else {
        filled.set(45);
      }

      byteQ3.add(b);

      return this;
    }

    public TraceBuilder byteR0(final UnsignedByte b) {
      if (filled.get(46)) {
        throw new IllegalStateException("BYTE_R_0 already set");
      } else {
        filled.set(46);
      }

      byteR0.add(b);

      return this;
    }

    public TraceBuilder byteR1(final UnsignedByte b) {
      if (filled.get(47)) {
        throw new IllegalStateException("BYTE_R_1 already set");
      } else {
        filled.set(47);
      }

      byteR1.add(b);

      return this;
    }

    public TraceBuilder byteR2(final UnsignedByte b) {
      if (filled.get(48)) {
        throw new IllegalStateException("BYTE_R_2 already set");
      } else {
        filled.set(48);
      }

      byteR2.add(b);

      return this;
    }

    public TraceBuilder byteR3(final UnsignedByte b) {
      if (filled.get(49)) {
        throw new IllegalStateException("BYTE_R_3 already set");
      } else {
        filled.set(49);
      }

      byteR3.add(b);

      return this;
    }

    public TraceBuilder cmp1(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("CMP_1 already set");
      } else {
        filled.set(50);
      }

      cmp1.add(b);

      return this;
    }

    public TraceBuilder cmp2(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("CMP_2 already set");
      } else {
        filled.set(51);
      }

      cmp2.add(b);

      return this;
    }

    public TraceBuilder ct(final BigInteger b) {
      if (filled.get(52)) {
        throw new IllegalStateException("CT already set");
      } else {
        filled.set(52);
      }

      ct.add(b);

      return this;
    }

    public TraceBuilder decOutput(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("DEC_OUTPUT already set");
      } else {
        filled.set(53);
      }

      decOutput.add(b);

      return this;
    }

    public TraceBuilder decSigned(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("DEC_SIGNED already set");
      } else {
        filled.set(54);
      }

      decSigned.add(b);

      return this;
    }

    public TraceBuilder inst(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(55);
      }

      inst.add(b);

      return this;
    }

    public TraceBuilder msb1(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("MSB_1 already set");
      } else {
        filled.set(56);
      }

      msb1.add(b);

      return this;
    }

    public TraceBuilder msb2(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("MSB_2 already set");
      } else {
        filled.set(57);
      }

      msb2.add(b);

      return this;
    }

    public TraceBuilder oli(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("OLI already set");
      } else {
        filled.set(58);
      }

      oli.add(b);

      return this;
    }

    public TraceBuilder resHi(final BigInteger b) {
      if (filled.get(59)) {
        throw new IllegalStateException("RES_HI already set");
      } else {
        filled.set(59);
      }

      resHi.add(b);

      return this;
    }

    public TraceBuilder resLo(final BigInteger b) {
      if (filled.get(60)) {
        throw new IllegalStateException("RES_LO already set");
      } else {
        filled.set(60);
      }

      resLo.add(b);

      return this;
    }

    public TraceBuilder stamp(final BigInteger b) {
      if (filled.get(61)) {
        throw new IllegalStateException("STAMP already set");
      } else {
        filled.set(61);
      }

      stamp.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ACC_1_2 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_1_3 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ACC_2_2 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_2_3 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_B_0 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("ACC_B_1 has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("ACC_B_2 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ACC_B_3 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ACC_DELTA_0 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ACC_DELTA_1 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("ACC_DELTA_2 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("ACC_DELTA_3 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("ACC_H_0 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("ACC_H_1 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("ACC_H_2 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("ACC_Q_0 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("ACC_Q_1 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("ACC_Q_2 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("ACC_Q_3 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("ACC_R_0 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("ACC_R_1 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("ACC_R_2 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("ACC_R_3 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("ARG_1_HI has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("ARG_1_LO has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ARG_2_HI has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("ARG_2_LO has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("BYTE_1_2 has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("BYTE_1_3 has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("BYTE_2_2 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("BYTE_2_3 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("BYTE_B_0 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("BYTE_B_1 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("BYTE_B_2 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("BYTE_B_3 has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("BYTE_DELTA_0 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("BYTE_DELTA_1 has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("BYTE_DELTA_2 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("BYTE_DELTA_3 has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("BYTE_H_0 has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("BYTE_H_1 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("BYTE_H_2 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("BYTE_Q_0 has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("BYTE_Q_1 has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("BYTE_Q_2 has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("BYTE_Q_3 has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("BYTE_R_0 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("BYTE_R_1 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("BYTE_R_2 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("BYTE_R_3 has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("CMP_1 has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("CMP_2 has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("CT has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("DEC_OUTPUT has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("DEC_SIGNED has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("MSB_1 has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("MSB_2 has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("OLI has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("RES_HI has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("RES_LO has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException("STAMP has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        acc12.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        acc13.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        acc22.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        acc23.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        accB0.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        accB1.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        accB2.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        accB3.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        accDelta0.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        accDelta1.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        accDelta2.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        accDelta3.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        accH0.add(BigInteger.ZERO);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        accH1.add(BigInteger.ZERO);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        accH2.add(BigInteger.ZERO);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        accQ0.add(BigInteger.ZERO);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        accQ1.add(BigInteger.ZERO);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        accQ2.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        accQ3.add(BigInteger.ZERO);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        accR0.add(BigInteger.ZERO);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        accR1.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        accR2.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        accR3.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        arg1Hi.add(BigInteger.ZERO);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        arg1Lo.add(BigInteger.ZERO);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        arg2Hi.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        arg2Lo.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        byte12.add(UnsignedByte.of(0));
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        byte13.add(UnsignedByte.of(0));
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        byte22.add(UnsignedByte.of(0));
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        byte23.add(UnsignedByte.of(0));
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        byteB0.add(UnsignedByte.of(0));
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        byteB1.add(UnsignedByte.of(0));
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        byteB2.add(UnsignedByte.of(0));
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        byteB3.add(UnsignedByte.of(0));
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        byteDelta0.add(UnsignedByte.of(0));
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        byteDelta1.add(UnsignedByte.of(0));
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        byteDelta2.add(UnsignedByte.of(0));
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        byteDelta3.add(UnsignedByte.of(0));
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        byteH0.add(UnsignedByte.of(0));
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        byteH1.add(UnsignedByte.of(0));
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        byteH2.add(UnsignedByte.of(0));
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        byteQ0.add(UnsignedByte.of(0));
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        byteQ1.add(UnsignedByte.of(0));
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        byteQ2.add(UnsignedByte.of(0));
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        byteQ3.add(UnsignedByte.of(0));
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        byteR0.add(UnsignedByte.of(0));
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        byteR1.add(UnsignedByte.of(0));
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        byteR2.add(UnsignedByte.of(0));
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        byteR3.add(UnsignedByte.of(0));
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        cmp1.add(false);
        this.filled.set(50);
      }
      if (!filled.get(51)) {
        cmp2.add(false);
        this.filled.set(51);
      }
      if (!filled.get(52)) {
        ct.add(BigInteger.ZERO);
        this.filled.set(52);
      }
      if (!filled.get(53)) {
        decOutput.add(false);
        this.filled.set(53);
      }
      if (!filled.get(54)) {
        decSigned.add(false);
        this.filled.set(54);
      }
      if (!filled.get(55)) {
        inst.add(BigInteger.ZERO);
        this.filled.set(55);
      }
      if (!filled.get(56)) {
        msb1.add(false);
        this.filled.set(56);
      }
      if (!filled.get(57)) {
        msb2.add(false);
        this.filled.set(57);
      }
      if (!filled.get(58)) {
        oli.add(false);
        this.filled.set(58);
      }
      if (!filled.get(59)) {
        resHi.add(BigInteger.ZERO);
        this.filled.set(59);
      }
      if (!filled.get(60)) {
        resLo.add(BigInteger.ZERO);
        this.filled.set(60);
      }
      if (!filled.get(61)) {
        stamp.add(BigInteger.ZERO);
        this.filled.set(61);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          acc12,
          acc13,
          acc22,
          acc23,
          accB0,
          accB1,
          accB2,
          accB3,
          accDelta0,
          accDelta1,
          accDelta2,
          accDelta3,
          accH0,
          accH1,
          accH2,
          accQ0,
          accQ1,
          accQ2,
          accQ3,
          accR0,
          accR1,
          accR2,
          accR3,
          arg1Hi,
          arg1Lo,
          arg2Hi,
          arg2Lo,
          byte12,
          byte13,
          byte22,
          byte23,
          byteB0,
          byteB1,
          byteB2,
          byteB3,
          byteDelta0,
          byteDelta1,
          byteDelta2,
          byteDelta3,
          byteH0,
          byteH1,
          byteH2,
          byteQ0,
          byteQ1,
          byteQ2,
          byteQ3,
          byteR0,
          byteR1,
          byteR2,
          byteR3,
          cmp1,
          cmp2,
          ct,
          decOutput,
          decSigned,
          inst,
          msb1,
          msb2,
          oli,
          resHi,
          resLo,
          stamp);
    }
  }
}
