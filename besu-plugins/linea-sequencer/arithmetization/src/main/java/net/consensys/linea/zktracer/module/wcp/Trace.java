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

package net.consensys.linea.zktracer.module.wcp;

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
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ACC_3") List<BigInteger> acc3,
    @JsonProperty("ACC_4") List<BigInteger> acc4,
    @JsonProperty("ACC_5") List<BigInteger> acc5,
    @JsonProperty("ACC_6") List<BigInteger> acc6,
    @JsonProperty("ARGUMENT_1_HI") List<BigInteger> argument1Hi,
    @JsonProperty("ARGUMENT_1_LO") List<BigInteger> argument1Lo,
    @JsonProperty("ARGUMENT_2_HI") List<BigInteger> argument2Hi,
    @JsonProperty("ARGUMENT_2_LO") List<BigInteger> argument2Lo,
    @JsonProperty("BIT_1") List<Boolean> bit1,
    @JsonProperty("BIT_2") List<Boolean> bit2,
    @JsonProperty("BIT_3") List<Boolean> bit3,
    @JsonProperty("BIT_4") List<Boolean> bit4,
    @JsonProperty("BITS") List<Boolean> bits,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("BYTE_3") List<UnsignedByte> byte3,
    @JsonProperty("BYTE_4") List<UnsignedByte> byte4,
    @JsonProperty("BYTE_5") List<UnsignedByte> byte5,
    @JsonProperty("BYTE_6") List<UnsignedByte> byte6,
    @JsonProperty("COUNTER") List<BigInteger> counter,
    @JsonProperty("INST") List<BigInteger> inst,
    @JsonProperty("NEG_1") List<Boolean> neg1,
    @JsonProperty("NEG_2") List<Boolean> neg2,
    @JsonProperty("ONE_LINE_INSTRUCTION") List<Boolean> oneLineInstruction,
    @JsonProperty("RESULT_HI") List<BigInteger> resultHi,
    @JsonProperty("RESULT_LO") List<BigInteger> resultLo,
    @JsonProperty("WORD_COMPARISON_STAMP") List<BigInteger> wordComparisonStamp) {
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.acc1.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ACC_1")
    private final List<BigInteger> acc1;

    @JsonProperty("ACC_2")
    private final List<BigInteger> acc2;

    @JsonProperty("ACC_3")
    private final List<BigInteger> acc3;

    @JsonProperty("ACC_4")
    private final List<BigInteger> acc4;

    @JsonProperty("ACC_5")
    private final List<BigInteger> acc5;

    @JsonProperty("ACC_6")
    private final List<BigInteger> acc6;

    @JsonProperty("ARGUMENT_1_HI")
    private final List<BigInteger> argument1Hi;

    @JsonProperty("ARGUMENT_1_LO")
    private final List<BigInteger> argument1Lo;

    @JsonProperty("ARGUMENT_2_HI")
    private final List<BigInteger> argument2Hi;

    @JsonProperty("ARGUMENT_2_LO")
    private final List<BigInteger> argument2Lo;

    @JsonProperty("BIT_1")
    private final List<Boolean> bit1;

    @JsonProperty("BIT_2")
    private final List<Boolean> bit2;

    @JsonProperty("BIT_3")
    private final List<Boolean> bit3;

    @JsonProperty("BIT_4")
    private final List<Boolean> bit4;

    @JsonProperty("BITS")
    private final List<Boolean> bits;

    @JsonProperty("BYTE_1")
    private final List<UnsignedByte> byte1;

    @JsonProperty("BYTE_2")
    private final List<UnsignedByte> byte2;

    @JsonProperty("BYTE_3")
    private final List<UnsignedByte> byte3;

    @JsonProperty("BYTE_4")
    private final List<UnsignedByte> byte4;

    @JsonProperty("BYTE_5")
    private final List<UnsignedByte> byte5;

    @JsonProperty("BYTE_6")
    private final List<UnsignedByte> byte6;

    @JsonProperty("COUNTER")
    private final List<BigInteger> counter;

    @JsonProperty("INST")
    private final List<BigInteger> inst;

    @JsonProperty("NEG_1")
    private final List<Boolean> neg1;

    @JsonProperty("NEG_2")
    private final List<Boolean> neg2;

    @JsonProperty("ONE_LINE_INSTRUCTION")
    private final List<Boolean> oneLineInstruction;

    @JsonProperty("RESULT_HI")
    private final List<BigInteger> resultHi;

    @JsonProperty("RESULT_LO")
    private final List<BigInteger> resultLo;

    @JsonProperty("WORD_COMPARISON_STAMP")
    private final List<BigInteger> wordComparisonStamp;

    private TraceBuilder(int length) {
      this.acc1 = new ArrayList<>(length);
      this.acc2 = new ArrayList<>(length);
      this.acc3 = new ArrayList<>(length);
      this.acc4 = new ArrayList<>(length);
      this.acc5 = new ArrayList<>(length);
      this.acc6 = new ArrayList<>(length);
      this.argument1Hi = new ArrayList<>(length);
      this.argument1Lo = new ArrayList<>(length);
      this.argument2Hi = new ArrayList<>(length);
      this.argument2Lo = new ArrayList<>(length);
      this.bit1 = new ArrayList<>(length);
      this.bit2 = new ArrayList<>(length);
      this.bit3 = new ArrayList<>(length);
      this.bit4 = new ArrayList<>(length);
      this.bits = new ArrayList<>(length);
      this.byte1 = new ArrayList<>(length);
      this.byte2 = new ArrayList<>(length);
      this.byte3 = new ArrayList<>(length);
      this.byte4 = new ArrayList<>(length);
      this.byte5 = new ArrayList<>(length);
      this.byte6 = new ArrayList<>(length);
      this.counter = new ArrayList<>(length);
      this.inst = new ArrayList<>(length);
      this.neg1 = new ArrayList<>(length);
      this.neg2 = new ArrayList<>(length);
      this.oneLineInstruction = new ArrayList<>(length);
      this.resultHi = new ArrayList<>(length);
      this.resultLo = new ArrayList<>(length);
      this.wordComparisonStamp = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.acc1.size();
    }

    public TraceBuilder acc1(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(0);
      }

      acc1.add(b);

      return this;
    }

    public TraceBuilder acc2(final BigInteger b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(1);
      }

      acc2.add(b);

      return this;
    }

    public TraceBuilder acc3(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("ACC_3 already set");
      } else {
        filled.set(2);
      }

      acc3.add(b);

      return this;
    }

    public TraceBuilder acc4(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("ACC_4 already set");
      } else {
        filled.set(3);
      }

      acc4.add(b);

      return this;
    }

    public TraceBuilder acc5(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_5 already set");
      } else {
        filled.set(4);
      }

      acc5.add(b);

      return this;
    }

    public TraceBuilder acc6(final BigInteger b) {
      if (filled.get(5)) {
        throw new IllegalStateException("ACC_6 already set");
      } else {
        filled.set(5);
      }

      acc6.add(b);

      return this;
    }

    public TraceBuilder argument1Hi(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ARGUMENT_1_HI already set");
      } else {
        filled.set(6);
      }

      argument1Hi.add(b);

      return this;
    }

    public TraceBuilder argument1Lo(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ARGUMENT_1_LO already set");
      } else {
        filled.set(7);
      }

      argument1Lo.add(b);

      return this;
    }

    public TraceBuilder argument2Hi(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ARGUMENT_2_HI already set");
      } else {
        filled.set(8);
      }

      argument2Hi.add(b);

      return this;
    }

    public TraceBuilder argument2Lo(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ARGUMENT_2_LO already set");
      } else {
        filled.set(9);
      }

      argument2Lo.add(b);

      return this;
    }

    public TraceBuilder bit1(final Boolean b) {
      if (filled.get(11)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(11);
      }

      bit1.add(b);

      return this;
    }

    public TraceBuilder bit2(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(12);
      }

      bit2.add(b);

      return this;
    }

    public TraceBuilder bit3(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(13);
      }

      bit3.add(b);

      return this;
    }

    public TraceBuilder bit4(final Boolean b) {
      if (filled.get(14)) {
        throw new IllegalStateException("BIT_4 already set");
      } else {
        filled.set(14);
      }

      bit4.add(b);

      return this;
    }

    public TraceBuilder bits(final Boolean b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BITS already set");
      } else {
        filled.set(10);
      }

      bits.add(b);

      return this;
    }

    public TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(15);
      }

      byte1.add(b);

      return this;
    }

    public TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(16)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(16);
      }

      byte2.add(b);

      return this;
    }

    public TraceBuilder byte3(final UnsignedByte b) {
      if (filled.get(17)) {
        throw new IllegalStateException("BYTE_3 already set");
      } else {
        filled.set(17);
      }

      byte3.add(b);

      return this;
    }

    public TraceBuilder byte4(final UnsignedByte b) {
      if (filled.get(18)) {
        throw new IllegalStateException("BYTE_4 already set");
      } else {
        filled.set(18);
      }

      byte4.add(b);

      return this;
    }

    public TraceBuilder byte5(final UnsignedByte b) {
      if (filled.get(19)) {
        throw new IllegalStateException("BYTE_5 already set");
      } else {
        filled.set(19);
      }

      byte5.add(b);

      return this;
    }

    public TraceBuilder byte6(final UnsignedByte b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BYTE_6 already set");
      } else {
        filled.set(20);
      }

      byte6.add(b);

      return this;
    }

    public TraceBuilder counter(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(21);
      }

      counter.add(b);

      return this;
    }

    public TraceBuilder inst(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(22);
      }

      inst.add(b);

      return this;
    }

    public TraceBuilder neg1(final Boolean b) {
      if (filled.get(23)) {
        throw new IllegalStateException("NEG_1 already set");
      } else {
        filled.set(23);
      }

      neg1.add(b);

      return this;
    }

    public TraceBuilder neg2(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("NEG_2 already set");
      } else {
        filled.set(24);
      }

      neg2.add(b);

      return this;
    }

    public TraceBuilder oneLineInstruction(final Boolean b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION already set");
      } else {
        filled.set(25);
      }

      oneLineInstruction.add(b);

      return this;
    }

    public TraceBuilder resultHi(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("RESULT_HI already set");
      } else {
        filled.set(26);
      }

      resultHi.add(b);

      return this;
    }

    public TraceBuilder resultLo(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("RESULT_LO already set");
      } else {
        filled.set(27);
      }

      resultLo.add(b);

      return this;
    }

    public TraceBuilder wordComparisonStamp(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("WORD_COMPARISON_STAMP already set");
      } else {
        filled.set(28);
      }

      wordComparisonStamp.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("ACC_3 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("ACC_4 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_5 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("ACC_6 has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("ARGUMENT_1_HI has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ARGUMENT_1_LO has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ARGUMENT_2_HI has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ARGUMENT_2_LO has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("BIT_4 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BITS has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("BYTE_3 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("BYTE_4 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("BYTE_5 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BYTE_6 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("NEG_1 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("NEG_2 has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("RESULT_HI has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("RESULT_LO has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("WORD_COMPARISON_STAMP has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        acc1.add(BigInteger.ZERO);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        acc2.add(BigInteger.ZERO);
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        acc3.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        acc4.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        acc5.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        acc6.add(BigInteger.ZERO);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        argument1Hi.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        argument1Lo.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        argument2Hi.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        argument2Lo.add(BigInteger.ZERO);
        this.filled.set(9);
      }
      if (!filled.get(11)) {
        bit1.add(false);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        bit2.add(false);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        bit3.add(false);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        bit4.add(false);
        this.filled.set(14);
      }
      if (!filled.get(10)) {
        bits.add(false);
        this.filled.set(10);
      }
      if (!filled.get(15)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        byte2.add(UnsignedByte.of(0));
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        byte3.add(UnsignedByte.of(0));
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        byte4.add(UnsignedByte.of(0));
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        byte5.add(UnsignedByte.of(0));
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        byte6.add(UnsignedByte.of(0));
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        counter.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        inst.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        neg1.add(false);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        neg2.add(false);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        oneLineInstruction.add(false);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        resultHi.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        resultLo.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        wordComparisonStamp.add(BigInteger.ZERO);
        this.filled.set(28);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          acc1,
          acc2,
          acc3,
          acc4,
          acc5,
          acc6,
          argument1Hi,
          argument1Lo,
          argument2Hi,
          argument2Lo,
          bit1,
          bit2,
          bit3,
          bit4,
          bits,
          byte1,
          byte2,
          byte3,
          byte4,
          byte5,
          byte6,
          counter,
          inst,
          neg1,
          neg2,
          oneLineInstruction,
          resultHi,
          resultLo,
          wordComparisonStamp);
    }
  }
}
