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
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    private final List<BigInteger> acc1 = new ArrayList<>();
    private final List<BigInteger> acc2 = new ArrayList<>();
    private final List<BigInteger> acc3 = new ArrayList<>();
    private final List<BigInteger> acc4 = new ArrayList<>();
    private final List<BigInteger> acc5 = new ArrayList<>();
    private final List<BigInteger> acc6 = new ArrayList<>();
    private final List<BigInteger> argument1Hi = new ArrayList<>();
    private final List<BigInteger> argument1Lo = new ArrayList<>();
    private final List<BigInteger> argument2Hi = new ArrayList<>();
    private final List<BigInteger> argument2Lo = new ArrayList<>();
    private final List<Boolean> bit1 = new ArrayList<>();
    private final List<Boolean> bit2 = new ArrayList<>();
    private final List<Boolean> bit3 = new ArrayList<>();
    private final List<Boolean> bit4 = new ArrayList<>();
    private final List<Boolean> bits = new ArrayList<>();
    private final List<UnsignedByte> byte1 = new ArrayList<>();
    private final List<UnsignedByte> byte2 = new ArrayList<>();
    private final List<UnsignedByte> byte3 = new ArrayList<>();
    private final List<UnsignedByte> byte4 = new ArrayList<>();
    private final List<UnsignedByte> byte5 = new ArrayList<>();
    private final List<UnsignedByte> byte6 = new ArrayList<>();
    private final List<BigInteger> counter = new ArrayList<>();
    private final List<BigInteger> inst = new ArrayList<>();
    private final List<Boolean> neg1 = new ArrayList<>();
    private final List<Boolean> neg2 = new ArrayList<>();
    private final List<Boolean> oneLineInstruction = new ArrayList<>();
    private final List<BigInteger> resultHi = new ArrayList<>();
    private final List<BigInteger> resultLo = new ArrayList<>();
    private final List<BigInteger> wordComparisonStamp = new ArrayList<>();

    private TraceBuilder() {}

    TraceBuilder acc1(final BigInteger b) {
      if (filled.get(18)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(18);
      }

      acc1.add(b);

      return this;
    }

    TraceBuilder acc2(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(10);
      }

      acc2.add(b);

      return this;
    }

    TraceBuilder acc3(final BigInteger b) {
      if (filled.get(22)) {
        throw new IllegalStateException("ACC_3 already set");
      } else {
        filled.set(22);
      }

      acc3.add(b);

      return this;
    }

    TraceBuilder acc4(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("ACC_4 already set");
      } else {
        filled.set(4);
      }

      acc4.add(b);

      return this;
    }

    TraceBuilder acc5(final BigInteger b) {
      if (filled.get(12)) {
        throw new IllegalStateException("ACC_5 already set");
      } else {
        filled.set(12);
      }

      acc5.add(b);

      return this;
    }

    TraceBuilder acc6(final BigInteger b) {
      if (filled.get(13)) {
        throw new IllegalStateException("ACC_6 already set");
      } else {
        filled.set(13);
      }

      acc6.add(b);

      return this;
    }

    TraceBuilder argument1Hi(final BigInteger b) {
      if (filled.get(9)) {
        throw new IllegalStateException("ARGUMENT_1_HI already set");
      } else {
        filled.set(9);
      }

      argument1Hi.add(b);

      return this;
    }

    TraceBuilder argument1Lo(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ARGUMENT_1_LO already set");
      } else {
        filled.set(7);
      }

      argument1Lo.add(b);

      return this;
    }

    TraceBuilder argument2Hi(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("ARGUMENT_2_HI already set");
      } else {
        filled.set(25);
      }

      argument2Hi.add(b);

      return this;
    }

    TraceBuilder argument2Lo(final BigInteger b) {
      if (filled.get(24)) {
        throw new IllegalStateException("ARGUMENT_2_LO already set");
      } else {
        filled.set(24);
      }

      argument2Lo.add(b);

      return this;
    }

    TraceBuilder bits(final Boolean b) {
      if (filled.get(26)) {
        throw new IllegalStateException("BITS already set");
      } else {
        filled.set(26);
      }

      bits.add(b);

      return this;
    }

    TraceBuilder bit1(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(8);
      }

      bit1.add(b);

      return this;
    }

    TraceBuilder bit2(final Boolean b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(15);
      }

      bit2.add(b);

      return this;
    }

    TraceBuilder bit3(final Boolean b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(20);
      }

      bit3.add(b);

      return this;
    }

    TraceBuilder bit4(final Boolean b) {
      if (filled.get(5)) {
        throw new IllegalStateException("BIT_4 already set");
      } else {
        filled.set(5);
      }

      bit4.add(b);

      return this;
    }

    TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(28)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(28);
      }

      byte1.add(b);

      return this;
    }

    TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(27)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(27);
      }

      byte2.add(b);

      return this;
    }

    TraceBuilder byte3(final UnsignedByte b) {
      if (filled.get(11)) {
        throw new IllegalStateException("BYTE_3 already set");
      } else {
        filled.set(11);
      }

      byte3.add(b);

      return this;
    }

    TraceBuilder byte4(final UnsignedByte b) {
      if (filled.get(3)) {
        throw new IllegalStateException("BYTE_4 already set");
      } else {
        filled.set(3);
      }

      byte4.add(b);

      return this;
    }

    TraceBuilder byte5(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BYTE_5 already set");
      } else {
        filled.set(23);
      }

      byte5.add(b);

      return this;
    }

    TraceBuilder byte6(final UnsignedByte b) {
      if (filled.get(17)) {
        throw new IllegalStateException("BYTE_6 already set");
      } else {
        filled.set(17);
      }

      byte6.add(b);

      return this;
    }

    TraceBuilder counter(final BigInteger b) {
      if (filled.get(14)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(14);
      }

      counter.add(b);

      return this;
    }

    TraceBuilder inst(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("INST already set");
      } else {
        filled.set(6);
      }

      inst.add(b);

      return this;
    }

    TraceBuilder neg1(final Boolean b) {
      if (filled.get(1)) {
        throw new IllegalStateException("NEG_1 already set");
      } else {
        filled.set(1);
      }

      neg1.add(b);

      return this;
    }

    TraceBuilder neg2(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("NEG_2 already set");
      } else {
        filled.set(19);
      }

      neg2.add(b);

      return this;
    }

    TraceBuilder oneLineInstruction(final Boolean b) {
      if (filled.get(21)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION already set");
      } else {
        filled.set(21);
      }

      oneLineInstruction.add(b);

      return this;
    }

    TraceBuilder resultHi(final BigInteger b) {
      if (filled.get(16)) {
        throw new IllegalStateException("RESULT_HI already set");
      } else {
        filled.set(16);
      }

      resultHi.add(b);

      return this;
    }

    TraceBuilder resultLo(final BigInteger b) {
      if (filled.get(0)) {
        throw new IllegalStateException("RESULT_LO already set");
      } else {
        filled.set(0);
      }

      resultLo.add(b);

      return this;
    }

    TraceBuilder wordComparisonStamp(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("WORD_COMPARISON_STAMP already set");
      } else {
        filled.set(2);
      }

      wordComparisonStamp.add(b);

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

    TraceBuilder setAcc3At(final BigInteger b, int i) {
      acc3.set(i, b);

      return this;
    }

    TraceBuilder setAcc4At(final BigInteger b, int i) {
      acc4.set(i, b);

      return this;
    }

    TraceBuilder setAcc5At(final BigInteger b, int i) {
      acc5.set(i, b);

      return this;
    }

    TraceBuilder setAcc6At(final BigInteger b, int i) {
      acc6.set(i, b);

      return this;
    }

    TraceBuilder setArgument1HiAt(final BigInteger b, int i) {
      argument1Hi.set(i, b);

      return this;
    }

    TraceBuilder setArgument1LoAt(final BigInteger b, int i) {
      argument1Lo.set(i, b);

      return this;
    }

    TraceBuilder setArgument2HiAt(final BigInteger b, int i) {
      argument2Hi.set(i, b);

      return this;
    }

    TraceBuilder setArgument2LoAt(final BigInteger b, int i) {
      argument2Lo.set(i, b);

      return this;
    }

    TraceBuilder setBitsAt(final Boolean b, int i) {
      bits.set(i, b);

      return this;
    }

    TraceBuilder setBit1At(final Boolean b, int i) {
      bit1.set(i, b);

      return this;
    }

    TraceBuilder setBit2At(final Boolean b, int i) {
      bit2.set(i, b);

      return this;
    }

    TraceBuilder setBit3At(final Boolean b, int i) {
      bit3.set(i, b);

      return this;
    }

    TraceBuilder setBit4At(final Boolean b, int i) {
      bit4.set(i, b);

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

    TraceBuilder setByte3At(final UnsignedByte b, int i) {
      byte3.set(i, b);

      return this;
    }

    TraceBuilder setByte4At(final UnsignedByte b, int i) {
      byte4.set(i, b);

      return this;
    }

    TraceBuilder setByte5At(final UnsignedByte b, int i) {
      byte5.set(i, b);

      return this;
    }

    TraceBuilder setByte6At(final UnsignedByte b, int i) {
      byte6.set(i, b);

      return this;
    }

    TraceBuilder setCounterAt(final BigInteger b, int i) {
      counter.set(i, b);

      return this;
    }

    TraceBuilder setInstAt(final BigInteger b, int i) {
      inst.set(i, b);

      return this;
    }

    TraceBuilder setNeg1At(final Boolean b, int i) {
      neg1.set(i, b);

      return this;
    }

    TraceBuilder setNeg2At(final Boolean b, int i) {
      neg2.set(i, b);

      return this;
    }

    TraceBuilder setOneLineInstructionAt(final Boolean b, int i) {
      oneLineInstruction.set(i, b);

      return this;
    }

    TraceBuilder setResultHiAt(final BigInteger b, int i) {
      resultHi.set(i, b);

      return this;
    }

    TraceBuilder setResultLoAt(final BigInteger b, int i) {
      resultLo.set(i, b);

      return this;
    }

    TraceBuilder setWordComparisonStampAt(final BigInteger b, int i) {
      wordComparisonStamp.set(i, b);

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

    TraceBuilder setAcc3Relative(final BigInteger b, int i) {
      acc3.set(acc3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc4Relative(final BigInteger b, int i) {
      acc4.set(acc4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc5Relative(final BigInteger b, int i) {
      acc5.set(acc5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAcc6Relative(final BigInteger b, int i) {
      acc6.set(acc6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArgument1HiRelative(final BigInteger b, int i) {
      argument1Hi.set(argument1Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArgument1LoRelative(final BigInteger b, int i) {
      argument1Lo.set(argument1Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArgument2HiRelative(final BigInteger b, int i) {
      argument2Hi.set(argument2Hi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setArgument2LoRelative(final BigInteger b, int i) {
      argument2Lo.set(argument2Lo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitsRelative(final Boolean b, int i) {
      bits.set(bits.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit1Relative(final Boolean b, int i) {
      bit1.set(bit1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit2Relative(final Boolean b, int i) {
      bit2.set(bit2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit3Relative(final Boolean b, int i) {
      bit3.set(bit3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBit4Relative(final Boolean b, int i) {
      bit4.set(bit4.size() - 1 - i, b);

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

    TraceBuilder setByte3Relative(final UnsignedByte b, int i) {
      byte3.set(byte3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte4Relative(final UnsignedByte b, int i) {
      byte4.set(byte4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte5Relative(final UnsignedByte b, int i) {
      byte5.set(byte5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setByte6Relative(final UnsignedByte b, int i) {
      byte6.set(byte6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterRelative(final BigInteger b, int i) {
      counter.set(counter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInstRelative(final BigInteger b, int i) {
      inst.set(inst.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNeg1Relative(final Boolean b, int i) {
      neg1.set(neg1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNeg2Relative(final Boolean b, int i) {
      neg2.set(neg2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setOneLineInstructionRelative(final Boolean b, int i) {
      oneLineInstruction.set(oneLineInstruction.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResultHiRelative(final BigInteger b, int i) {
      resultHi.set(resultHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setResultLoRelative(final BigInteger b, int i) {
      resultLo.set(resultLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setWordComparisonStampRelative(final BigInteger b, int i) {
      wordComparisonStamp.set(wordComparisonStamp.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(18)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("ACC_3 has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("ACC_4 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("ACC_5 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("ACC_6 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("ARGUMENT_1_HI has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ARGUMENT_1_LO has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("ARGUMENT_2_HI has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("ARGUMENT_2_LO has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("BIT_4 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("BITS has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("BYTE_3 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("BYTE_4 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BYTE_5 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("BYTE_6 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("INST has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("NEG_1 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("NEG_2 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("ONE_LINE_INSTRUCTION has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("RESULT_HI has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("RESULT_LO has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("WORD_COMPARISON_STAMP has not been filled");
      }

      filled.clear();

      return this;
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
