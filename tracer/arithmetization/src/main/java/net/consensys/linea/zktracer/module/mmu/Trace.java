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

package net.consensys.linea.zktracer.module.mmu;

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
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ACC_3") List<BigInteger> acc3,
    @JsonProperty("ACC_4") List<BigInteger> acc4,
    @JsonProperty("ACC_5") List<BigInteger> acc5,
    @JsonProperty("ACC_6") List<BigInteger> acc6,
    @JsonProperty("ACC_7") List<BigInteger> acc7,
    @JsonProperty("ACC_8") List<BigInteger> acc8,
    @JsonProperty("ALIGNED") List<BigInteger> aligned,
    @JsonProperty("BIT_1") List<Boolean> bit1,
    @JsonProperty("BIT_2") List<Boolean> bit2,
    @JsonProperty("BIT_3") List<Boolean> bit3,
    @JsonProperty("BIT_4") List<Boolean> bit4,
    @JsonProperty("BIT_5") List<Boolean> bit5,
    @JsonProperty("BIT_6") List<Boolean> bit6,
    @JsonProperty("BIT_7") List<Boolean> bit7,
    @JsonProperty("BIT_8") List<Boolean> bit8,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("BYTE_3") List<UnsignedByte> byte3,
    @JsonProperty("BYTE_4") List<UnsignedByte> byte4,
    @JsonProperty("BYTE_5") List<UnsignedByte> byte5,
    @JsonProperty("BYTE_6") List<UnsignedByte> byte6,
    @JsonProperty("BYTE_7") List<UnsignedByte> byte7,
    @JsonProperty("BYTE_8") List<UnsignedByte> byte8,
    @JsonProperty("CALL_DATA_OFFSET") List<BigInteger> callDataOffset,
    @JsonProperty("CALL_DATA_SIZE") List<BigInteger> callDataSize,
    @JsonProperty("CALL_STACK_DEPTH") List<BigInteger> callStackDepth,
    @JsonProperty("CALLER") List<BigInteger> caller,
    @JsonProperty("CONTEXT_NUMBER") List<BigInteger> contextNumber,
    @JsonProperty("CONTEXT_SOURCE") List<BigInteger> contextSource,
    @JsonProperty("CONTEXT_TARGET") List<BigInteger> contextTarget,
    @JsonProperty("COUNTER") List<BigInteger> counter,
    @JsonProperty("ERF") List<Boolean> erf,
    @JsonProperty("EXO_IS_HASH") List<Boolean> exoIsHash,
    @JsonProperty("EXO_IS_LOG") List<Boolean> exoIsLog,
    @JsonProperty("EXO_IS_ROM") List<Boolean> exoIsRom,
    @JsonProperty("EXO_IS_TXCD") List<Boolean> exoIsTxcd,
    @JsonProperty("FAST") List<BigInteger> fast,
    @JsonProperty("INFO") List<BigInteger> info,
    @JsonProperty("INSTRUCTION") List<BigInteger> instruction,
    @JsonProperty("IS_DATA") List<Boolean> isData,
    @JsonProperty("IS_MICRO_INSTRUCTION") List<Boolean> isMicroInstruction,
    @JsonProperty("MICRO_INSTRUCTION") List<BigInteger> microInstruction,
    @JsonProperty("MICRO_INSTRUCTION_STAMP") List<BigInteger> microInstructionStamp,
    @JsonProperty("MIN") List<BigInteger> min,
    @JsonProperty("NIB_1") List<UnsignedByte> nib1,
    @JsonProperty("NIB_2") List<UnsignedByte> nib2,
    @JsonProperty("NIB_3") List<UnsignedByte> nib3,
    @JsonProperty("NIB_4") List<UnsignedByte> nib4,
    @JsonProperty("NIB_5") List<UnsignedByte> nib5,
    @JsonProperty("NIB_6") List<UnsignedByte> nib6,
    @JsonProperty("NIB_7") List<UnsignedByte> nib7,
    @JsonProperty("NIB_8") List<UnsignedByte> nib8,
    @JsonProperty("NIB_9") List<UnsignedByte> nib9,
    @JsonProperty("OFF_1_LO") List<BigInteger> off1Lo,
    @JsonProperty("OFF_2_HI") List<BigInteger> off2Hi,
    @JsonProperty("OFF_2_LO") List<BigInteger> off2Lo,
    @JsonProperty("OFFSET_OUT_OF_BOUNDS") List<Boolean> offsetOutOfBounds,
    @JsonProperty("PRECOMPUTATION") List<BigInteger> precomputation,
    @JsonProperty("RAM_STAMP") List<BigInteger> ramStamp,
    @JsonProperty("REFO") List<BigInteger> refo,
    @JsonProperty("REFS") List<BigInteger> refs,
    @JsonProperty("RETURN_CAPACITY") List<BigInteger> returnCapacity,
    @JsonProperty("RETURN_OFFSET") List<BigInteger> returnOffset,
    @JsonProperty("RETURNER") List<BigInteger> returner,
    @JsonProperty("SIZE") List<BigInteger> size_,
    @JsonProperty("SIZE_IMPORTED") List<BigInteger> sizeImported,
    @JsonProperty("SOURCE_BYTE_OFFSET") List<BigInteger> sourceByteOffset,
    @JsonProperty("SOURCE_LIMB_OFFSET") List<BigInteger> sourceLimbOffset,
    @JsonProperty("TARGET_BYTE_OFFSET") List<BigInteger> targetByteOffset,
    @JsonProperty("TARGET_LIMB_OFFSET") List<BigInteger> targetLimbOffset,
    @JsonProperty("TERNARY") List<BigInteger> ternary,
    @JsonProperty("TO_RAM") List<Boolean> toRam,
    @JsonProperty("TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS")
        List<BigInteger> totalNumberOfMicroInstructions,
    @JsonProperty("TOTAL_NUMBER_OF_PADDINGS") List<BigInteger> totalNumberOfPaddings,
    @JsonProperty("TOTAL_NUMBER_OF_READS") List<BigInteger> totalNumberOfReads,
    @JsonProperty("VAL_HI") List<BigInteger> valHi,
    @JsonProperty("VAL_LO") List<BigInteger> valLo) {
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

    @JsonProperty("ACC_7")
    private final List<BigInteger> acc7;

    @JsonProperty("ACC_8")
    private final List<BigInteger> acc8;

    @JsonProperty("ALIGNED")
    private final List<BigInteger> aligned;

    @JsonProperty("BIT_1")
    private final List<Boolean> bit1;

    @JsonProperty("BIT_2")
    private final List<Boolean> bit2;

    @JsonProperty("BIT_3")
    private final List<Boolean> bit3;

    @JsonProperty("BIT_4")
    private final List<Boolean> bit4;

    @JsonProperty("BIT_5")
    private final List<Boolean> bit5;

    @JsonProperty("BIT_6")
    private final List<Boolean> bit6;

    @JsonProperty("BIT_7")
    private final List<Boolean> bit7;

    @JsonProperty("BIT_8")
    private final List<Boolean> bit8;

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

    @JsonProperty("BYTE_7")
    private final List<UnsignedByte> byte7;

    @JsonProperty("BYTE_8")
    private final List<UnsignedByte> byte8;

    @JsonProperty("CALL_DATA_OFFSET")
    private final List<BigInteger> callDataOffset;

    @JsonProperty("CALL_DATA_SIZE")
    private final List<BigInteger> callDataSize;

    @JsonProperty("CALL_STACK_DEPTH")
    private final List<BigInteger> callStackDepth;

    @JsonProperty("CALLER")
    private final List<BigInteger> caller;

    @JsonProperty("CONTEXT_NUMBER")
    private final List<BigInteger> contextNumber;

    @JsonProperty("CONTEXT_SOURCE")
    private final List<BigInteger> contextSource;

    @JsonProperty("CONTEXT_TARGET")
    private final List<BigInteger> contextTarget;

    @JsonProperty("COUNTER")
    private final List<BigInteger> counter;

    @JsonProperty("ERF")
    private final List<Boolean> erf;

    @JsonProperty("EXO_IS_HASH")
    private final List<Boolean> exoIsHash;

    @JsonProperty("EXO_IS_LOG")
    private final List<Boolean> exoIsLog;

    @JsonProperty("EXO_IS_ROM")
    private final List<Boolean> exoIsRom;

    @JsonProperty("EXO_IS_TXCD")
    private final List<Boolean> exoIsTxcd;

    @JsonProperty("FAST")
    private final List<BigInteger> fast;

    @JsonProperty("INFO")
    private final List<BigInteger> info;

    @JsonProperty("INSTRUCTION")
    private final List<BigInteger> instruction;

    @JsonProperty("IS_DATA")
    private final List<Boolean> isData;

    @JsonProperty("IS_MICRO_INSTRUCTION")
    private final List<Boolean> isMicroInstruction;

    @JsonProperty("MICRO_INSTRUCTION")
    private final List<BigInteger> microInstruction;

    @JsonProperty("MICRO_INSTRUCTION_STAMP")
    private final List<BigInteger> microInstructionStamp;

    @JsonProperty("MIN")
    private final List<BigInteger> min;

    @JsonProperty("NIB_1")
    private final List<UnsignedByte> nib1;

    @JsonProperty("NIB_2")
    private final List<UnsignedByte> nib2;

    @JsonProperty("NIB_3")
    private final List<UnsignedByte> nib3;

    @JsonProperty("NIB_4")
    private final List<UnsignedByte> nib4;

    @JsonProperty("NIB_5")
    private final List<UnsignedByte> nib5;

    @JsonProperty("NIB_6")
    private final List<UnsignedByte> nib6;

    @JsonProperty("NIB_7")
    private final List<UnsignedByte> nib7;

    @JsonProperty("NIB_8")
    private final List<UnsignedByte> nib8;

    @JsonProperty("NIB_9")
    private final List<UnsignedByte> nib9;

    @JsonProperty("OFF_1_LO")
    private final List<BigInteger> off1Lo;

    @JsonProperty("OFF_2_HI")
    private final List<BigInteger> off2Hi;

    @JsonProperty("OFF_2_LO")
    private final List<BigInteger> off2Lo;

    @JsonProperty("OFFSET_OUT_OF_BOUNDS")
    private final List<Boolean> offsetOutOfBounds;

    @JsonProperty("PRECOMPUTATION")
    private final List<BigInteger> precomputation;

    @JsonProperty("RAM_STAMP")
    private final List<BigInteger> ramStamp;

    @JsonProperty("REFO")
    private final List<BigInteger> refo;

    @JsonProperty("REFS")
    private final List<BigInteger> refs;

    @JsonProperty("RETURN_CAPACITY")
    private final List<BigInteger> returnCapacity;

    @JsonProperty("RETURN_OFFSET")
    private final List<BigInteger> returnOffset;

    @JsonProperty("RETURNER")
    private final List<BigInteger> returner;

    @JsonProperty("SIZE")
    private final List<BigInteger> size;

    @JsonProperty("SIZE_IMPORTED")
    private final List<BigInteger> sizeImported;

    @JsonProperty("SOURCE_BYTE_OFFSET")
    private final List<BigInteger> sourceByteOffset;

    @JsonProperty("SOURCE_LIMB_OFFSET")
    private final List<BigInteger> sourceLimbOffset;

    @JsonProperty("TARGET_BYTE_OFFSET")
    private final List<BigInteger> targetByteOffset;

    @JsonProperty("TARGET_LIMB_OFFSET")
    private final List<BigInteger> targetLimbOffset;

    @JsonProperty("TERNARY")
    private final List<BigInteger> ternary;

    @JsonProperty("TO_RAM")
    private final List<Boolean> toRam;

    @JsonProperty("TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS")
    private final List<BigInteger> totalNumberOfMicroInstructions;

    @JsonProperty("TOTAL_NUMBER_OF_PADDINGS")
    private final List<BigInteger> totalNumberOfPaddings;

    @JsonProperty("TOTAL_NUMBER_OF_READS")
    private final List<BigInteger> totalNumberOfReads;

    @JsonProperty("VAL_HI")
    private final List<BigInteger> valHi;

    @JsonProperty("VAL_LO")
    private final List<BigInteger> valLo;

    private TraceBuilder(int length) {
      this.acc1 = new ArrayList<>(length);
      this.acc2 = new ArrayList<>(length);
      this.acc3 = new ArrayList<>(length);
      this.acc4 = new ArrayList<>(length);
      this.acc5 = new ArrayList<>(length);
      this.acc6 = new ArrayList<>(length);
      this.acc7 = new ArrayList<>(length);
      this.acc8 = new ArrayList<>(length);
      this.aligned = new ArrayList<>(length);
      this.bit1 = new ArrayList<>(length);
      this.bit2 = new ArrayList<>(length);
      this.bit3 = new ArrayList<>(length);
      this.bit4 = new ArrayList<>(length);
      this.bit5 = new ArrayList<>(length);
      this.bit6 = new ArrayList<>(length);
      this.bit7 = new ArrayList<>(length);
      this.bit8 = new ArrayList<>(length);
      this.byte1 = new ArrayList<>(length);
      this.byte2 = new ArrayList<>(length);
      this.byte3 = new ArrayList<>(length);
      this.byte4 = new ArrayList<>(length);
      this.byte5 = new ArrayList<>(length);
      this.byte6 = new ArrayList<>(length);
      this.byte7 = new ArrayList<>(length);
      this.byte8 = new ArrayList<>(length);
      this.callDataOffset = new ArrayList<>(length);
      this.callDataSize = new ArrayList<>(length);
      this.callStackDepth = new ArrayList<>(length);
      this.caller = new ArrayList<>(length);
      this.contextNumber = new ArrayList<>(length);
      this.contextSource = new ArrayList<>(length);
      this.contextTarget = new ArrayList<>(length);
      this.counter = new ArrayList<>(length);
      this.erf = new ArrayList<>(length);
      this.exoIsHash = new ArrayList<>(length);
      this.exoIsLog = new ArrayList<>(length);
      this.exoIsRom = new ArrayList<>(length);
      this.exoIsTxcd = new ArrayList<>(length);
      this.fast = new ArrayList<>(length);
      this.info = new ArrayList<>(length);
      this.instruction = new ArrayList<>(length);
      this.isData = new ArrayList<>(length);
      this.isMicroInstruction = new ArrayList<>(length);
      this.microInstruction = new ArrayList<>(length);
      this.microInstructionStamp = new ArrayList<>(length);
      this.min = new ArrayList<>(length);
      this.nib1 = new ArrayList<>(length);
      this.nib2 = new ArrayList<>(length);
      this.nib3 = new ArrayList<>(length);
      this.nib4 = new ArrayList<>(length);
      this.nib5 = new ArrayList<>(length);
      this.nib6 = new ArrayList<>(length);
      this.nib7 = new ArrayList<>(length);
      this.nib8 = new ArrayList<>(length);
      this.nib9 = new ArrayList<>(length);
      this.off1Lo = new ArrayList<>(length);
      this.off2Hi = new ArrayList<>(length);
      this.off2Lo = new ArrayList<>(length);
      this.offsetOutOfBounds = new ArrayList<>(length);
      this.precomputation = new ArrayList<>(length);
      this.ramStamp = new ArrayList<>(length);
      this.refo = new ArrayList<>(length);
      this.refs = new ArrayList<>(length);
      this.returnCapacity = new ArrayList<>(length);
      this.returnOffset = new ArrayList<>(length);
      this.returner = new ArrayList<>(length);
      this.size = new ArrayList<>(length);
      this.sizeImported = new ArrayList<>(length);
      this.sourceByteOffset = new ArrayList<>(length);
      this.sourceLimbOffset = new ArrayList<>(length);
      this.targetByteOffset = new ArrayList<>(length);
      this.targetLimbOffset = new ArrayList<>(length);
      this.ternary = new ArrayList<>(length);
      this.toRam = new ArrayList<>(length);
      this.totalNumberOfMicroInstructions = new ArrayList<>(length);
      this.totalNumberOfPaddings = new ArrayList<>(length);
      this.totalNumberOfReads = new ArrayList<>(length);
      this.valHi = new ArrayList<>(length);
      this.valLo = new ArrayList<>(length);
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

    public TraceBuilder acc7(final BigInteger b) {
      if (filled.get(6)) {
        throw new IllegalStateException("ACC_7 already set");
      } else {
        filled.set(6);
      }

      acc7.add(b);

      return this;
    }

    public TraceBuilder acc8(final BigInteger b) {
      if (filled.get(7)) {
        throw new IllegalStateException("ACC_8 already set");
      } else {
        filled.set(7);
      }

      acc8.add(b);

      return this;
    }

    public TraceBuilder aligned(final BigInteger b) {
      if (filled.get(8)) {
        throw new IllegalStateException("ALIGNED already set");
      } else {
        filled.set(8);
      }

      aligned.add(b);

      return this;
    }

    public TraceBuilder bit1(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("BIT_1 already set");
      } else {
        filled.set(9);
      }

      bit1.add(b);

      return this;
    }

    public TraceBuilder bit2(final Boolean b) {
      if (filled.get(10)) {
        throw new IllegalStateException("BIT_2 already set");
      } else {
        filled.set(10);
      }

      bit2.add(b);

      return this;
    }

    public TraceBuilder bit3(final Boolean b) {
      if (filled.get(11)) {
        throw new IllegalStateException("BIT_3 already set");
      } else {
        filled.set(11);
      }

      bit3.add(b);

      return this;
    }

    public TraceBuilder bit4(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("BIT_4 already set");
      } else {
        filled.set(12);
      }

      bit4.add(b);

      return this;
    }

    public TraceBuilder bit5(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("BIT_5 already set");
      } else {
        filled.set(13);
      }

      bit5.add(b);

      return this;
    }

    public TraceBuilder bit6(final Boolean b) {
      if (filled.get(14)) {
        throw new IllegalStateException("BIT_6 already set");
      } else {
        filled.set(14);
      }

      bit6.add(b);

      return this;
    }

    public TraceBuilder bit7(final Boolean b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BIT_7 already set");
      } else {
        filled.set(15);
      }

      bit7.add(b);

      return this;
    }

    public TraceBuilder bit8(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("BIT_8 already set");
      } else {
        filled.set(16);
      }

      bit8.add(b);

      return this;
    }

    public TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(17)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(17);
      }

      byte1.add(b);

      return this;
    }

    public TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(18)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(18);
      }

      byte2.add(b);

      return this;
    }

    public TraceBuilder byte3(final UnsignedByte b) {
      if (filled.get(19)) {
        throw new IllegalStateException("BYTE_3 already set");
      } else {
        filled.set(19);
      }

      byte3.add(b);

      return this;
    }

    public TraceBuilder byte4(final UnsignedByte b) {
      if (filled.get(20)) {
        throw new IllegalStateException("BYTE_4 already set");
      } else {
        filled.set(20);
      }

      byte4.add(b);

      return this;
    }

    public TraceBuilder byte5(final UnsignedByte b) {
      if (filled.get(21)) {
        throw new IllegalStateException("BYTE_5 already set");
      } else {
        filled.set(21);
      }

      byte5.add(b);

      return this;
    }

    public TraceBuilder byte6(final UnsignedByte b) {
      if (filled.get(22)) {
        throw new IllegalStateException("BYTE_6 already set");
      } else {
        filled.set(22);
      }

      byte6.add(b);

      return this;
    }

    public TraceBuilder byte7(final UnsignedByte b) {
      if (filled.get(23)) {
        throw new IllegalStateException("BYTE_7 already set");
      } else {
        filled.set(23);
      }

      byte7.add(b);

      return this;
    }

    public TraceBuilder byte8(final UnsignedByte b) {
      if (filled.get(24)) {
        throw new IllegalStateException("BYTE_8 already set");
      } else {
        filled.set(24);
      }

      byte8.add(b);

      return this;
    }

    public TraceBuilder callDataOffset(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("CALL_DATA_OFFSET already set");
      } else {
        filled.set(26);
      }

      callDataOffset.add(b);

      return this;
    }

    public TraceBuilder callDataSize(final BigInteger b) {
      if (filled.get(27)) {
        throw new IllegalStateException("CALL_DATA_SIZE already set");
      } else {
        filled.set(27);
      }

      callDataSize.add(b);

      return this;
    }

    public TraceBuilder callStackDepth(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("CALL_STACK_DEPTH already set");
      } else {
        filled.set(28);
      }

      callStackDepth.add(b);

      return this;
    }

    public TraceBuilder caller(final BigInteger b) {
      if (filled.get(25)) {
        throw new IllegalStateException("CALLER already set");
      } else {
        filled.set(25);
      }

      caller.add(b);

      return this;
    }

    public TraceBuilder contextNumber(final BigInteger b) {
      if (filled.get(29)) {
        throw new IllegalStateException("CONTEXT_NUMBER already set");
      } else {
        filled.set(29);
      }

      contextNumber.add(b);

      return this;
    }

    public TraceBuilder contextSource(final BigInteger b) {
      if (filled.get(30)) {
        throw new IllegalStateException("CONTEXT_SOURCE already set");
      } else {
        filled.set(30);
      }

      contextSource.add(b);

      return this;
    }

    public TraceBuilder contextTarget(final BigInteger b) {
      if (filled.get(31)) {
        throw new IllegalStateException("CONTEXT_TARGET already set");
      } else {
        filled.set(31);
      }

      contextTarget.add(b);

      return this;
    }

    public TraceBuilder counter(final BigInteger b) {
      if (filled.get(32)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(32);
      }

      counter.add(b);

      return this;
    }

    public TraceBuilder erf(final Boolean b) {
      if (filled.get(33)) {
        throw new IllegalStateException("ERF already set");
      } else {
        filled.set(33);
      }

      erf.add(b);

      return this;
    }

    public TraceBuilder exoIsHash(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("EXO_IS_HASH already set");
      } else {
        filled.set(34);
      }

      exoIsHash.add(b);

      return this;
    }

    public TraceBuilder exoIsLog(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("EXO_IS_LOG already set");
      } else {
        filled.set(35);
      }

      exoIsLog.add(b);

      return this;
    }

    public TraceBuilder exoIsRom(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("EXO_IS_ROM already set");
      } else {
        filled.set(36);
      }

      exoIsRom.add(b);

      return this;
    }

    public TraceBuilder exoIsTxcd(final Boolean b) {
      if (filled.get(37)) {
        throw new IllegalStateException("EXO_IS_TXCD already set");
      } else {
        filled.set(37);
      }

      exoIsTxcd.add(b);

      return this;
    }

    public TraceBuilder fast(final BigInteger b) {
      if (filled.get(38)) {
        throw new IllegalStateException("FAST already set");
      } else {
        filled.set(38);
      }

      fast.add(b);

      return this;
    }

    public TraceBuilder info(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("INFO already set");
      } else {
        filled.set(39);
      }

      info.add(b);

      return this;
    }

    public TraceBuilder instruction(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("INSTRUCTION already set");
      } else {
        filled.set(40);
      }

      instruction.add(b);

      return this;
    }

    public TraceBuilder isData(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("IS_DATA already set");
      } else {
        filled.set(41);
      }

      isData.add(b);

      return this;
    }

    public TraceBuilder isMicroInstruction(final Boolean b) {
      if (filled.get(42)) {
        throw new IllegalStateException("IS_MICRO_INSTRUCTION already set");
      } else {
        filled.set(42);
      }

      isMicroInstruction.add(b);

      return this;
    }

    public TraceBuilder microInstruction(final BigInteger b) {
      if (filled.get(43)) {
        throw new IllegalStateException("MICRO_INSTRUCTION already set");
      } else {
        filled.set(43);
      }

      microInstruction.add(b);

      return this;
    }

    public TraceBuilder microInstructionStamp(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("MICRO_INSTRUCTION_STAMP already set");
      } else {
        filled.set(44);
      }

      microInstructionStamp.add(b);

      return this;
    }

    public TraceBuilder min(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("MIN already set");
      } else {
        filled.set(45);
      }

      min.add(b);

      return this;
    }

    public TraceBuilder nib1(final UnsignedByte b) {
      if (filled.get(46)) {
        throw new IllegalStateException("NIB_1 already set");
      } else {
        filled.set(46);
      }

      nib1.add(b);

      return this;
    }

    public TraceBuilder nib2(final UnsignedByte b) {
      if (filled.get(47)) {
        throw new IllegalStateException("NIB_2 already set");
      } else {
        filled.set(47);
      }

      nib2.add(b);

      return this;
    }

    public TraceBuilder nib3(final UnsignedByte b) {
      if (filled.get(48)) {
        throw new IllegalStateException("NIB_3 already set");
      } else {
        filled.set(48);
      }

      nib3.add(b);

      return this;
    }

    public TraceBuilder nib4(final UnsignedByte b) {
      if (filled.get(49)) {
        throw new IllegalStateException("NIB_4 already set");
      } else {
        filled.set(49);
      }

      nib4.add(b);

      return this;
    }

    public TraceBuilder nib5(final UnsignedByte b) {
      if (filled.get(50)) {
        throw new IllegalStateException("NIB_5 already set");
      } else {
        filled.set(50);
      }

      nib5.add(b);

      return this;
    }

    public TraceBuilder nib6(final UnsignedByte b) {
      if (filled.get(51)) {
        throw new IllegalStateException("NIB_6 already set");
      } else {
        filled.set(51);
      }

      nib6.add(b);

      return this;
    }

    public TraceBuilder nib7(final UnsignedByte b) {
      if (filled.get(52)) {
        throw new IllegalStateException("NIB_7 already set");
      } else {
        filled.set(52);
      }

      nib7.add(b);

      return this;
    }

    public TraceBuilder nib8(final UnsignedByte b) {
      if (filled.get(53)) {
        throw new IllegalStateException("NIB_8 already set");
      } else {
        filled.set(53);
      }

      nib8.add(b);

      return this;
    }

    public TraceBuilder nib9(final UnsignedByte b) {
      if (filled.get(54)) {
        throw new IllegalStateException("NIB_9 already set");
      } else {
        filled.set(54);
      }

      nib9.add(b);

      return this;
    }

    public TraceBuilder off1Lo(final BigInteger b) {
      if (filled.get(56)) {
        throw new IllegalStateException("OFF_1_LO already set");
      } else {
        filled.set(56);
      }

      off1Lo.add(b);

      return this;
    }

    public TraceBuilder off2Hi(final BigInteger b) {
      if (filled.get(57)) {
        throw new IllegalStateException("OFF_2_HI already set");
      } else {
        filled.set(57);
      }

      off2Hi.add(b);

      return this;
    }

    public TraceBuilder off2Lo(final BigInteger b) {
      if (filled.get(58)) {
        throw new IllegalStateException("OFF_2_LO already set");
      } else {
        filled.set(58);
      }

      off2Lo.add(b);

      return this;
    }

    public TraceBuilder offsetOutOfBounds(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("OFFSET_OUT_OF_BOUNDS already set");
      } else {
        filled.set(55);
      }

      offsetOutOfBounds.add(b);

      return this;
    }

    public TraceBuilder precomputation(final BigInteger b) {
      if (filled.get(59)) {
        throw new IllegalStateException("PRECOMPUTATION already set");
      } else {
        filled.set(59);
      }

      precomputation.add(b);

      return this;
    }

    public TraceBuilder ramStamp(final BigInteger b) {
      if (filled.get(60)) {
        throw new IllegalStateException("RAM_STAMP already set");
      } else {
        filled.set(60);
      }

      ramStamp.add(b);

      return this;
    }

    public TraceBuilder refo(final BigInteger b) {
      if (filled.get(61)) {
        throw new IllegalStateException("REFO already set");
      } else {
        filled.set(61);
      }

      refo.add(b);

      return this;
    }

    public TraceBuilder refs(final BigInteger b) {
      if (filled.get(62)) {
        throw new IllegalStateException("REFS already set");
      } else {
        filled.set(62);
      }

      refs.add(b);

      return this;
    }

    public TraceBuilder returnCapacity(final BigInteger b) {
      if (filled.get(64)) {
        throw new IllegalStateException("RETURN_CAPACITY already set");
      } else {
        filled.set(64);
      }

      returnCapacity.add(b);

      return this;
    }

    public TraceBuilder returnOffset(final BigInteger b) {
      if (filled.get(65)) {
        throw new IllegalStateException("RETURN_OFFSET already set");
      } else {
        filled.set(65);
      }

      returnOffset.add(b);

      return this;
    }

    public TraceBuilder returner(final BigInteger b) {
      if (filled.get(63)) {
        throw new IllegalStateException("RETURNER already set");
      } else {
        filled.set(63);
      }

      returner.add(b);

      return this;
    }

    public TraceBuilder size(final BigInteger b) {
      if (filled.get(66)) {
        throw new IllegalStateException("SIZE already set");
      } else {
        filled.set(66);
      }

      size.add(b);

      return this;
    }

    public TraceBuilder sizeImported(final BigInteger b) {
      if (filled.get(67)) {
        throw new IllegalStateException("SIZE_IMPORTED already set");
      } else {
        filled.set(67);
      }

      sizeImported.add(b);

      return this;
    }

    public TraceBuilder sourceByteOffset(final BigInteger b) {
      if (filled.get(68)) {
        throw new IllegalStateException("SOURCE_BYTE_OFFSET already set");
      } else {
        filled.set(68);
      }

      sourceByteOffset.add(b);

      return this;
    }

    public TraceBuilder sourceLimbOffset(final BigInteger b) {
      if (filled.get(69)) {
        throw new IllegalStateException("SOURCE_LIMB_OFFSET already set");
      } else {
        filled.set(69);
      }

      sourceLimbOffset.add(b);

      return this;
    }

    public TraceBuilder targetByteOffset(final BigInteger b) {
      if (filled.get(70)) {
        throw new IllegalStateException("TARGET_BYTE_OFFSET already set");
      } else {
        filled.set(70);
      }

      targetByteOffset.add(b);

      return this;
    }

    public TraceBuilder targetLimbOffset(final BigInteger b) {
      if (filled.get(71)) {
        throw new IllegalStateException("TARGET_LIMB_OFFSET already set");
      } else {
        filled.set(71);
      }

      targetLimbOffset.add(b);

      return this;
    }

    public TraceBuilder ternary(final BigInteger b) {
      if (filled.get(72)) {
        throw new IllegalStateException("TERNARY already set");
      } else {
        filled.set(72);
      }

      ternary.add(b);

      return this;
    }

    public TraceBuilder toRam(final Boolean b) {
      if (filled.get(76)) {
        throw new IllegalStateException("TO_RAM already set");
      } else {
        filled.set(76);
      }

      toRam.add(b);

      return this;
    }

    public TraceBuilder totalNumberOfMicroInstructions(final BigInteger b) {
      if (filled.get(73)) {
        throw new IllegalStateException("TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS already set");
      } else {
        filled.set(73);
      }

      totalNumberOfMicroInstructions.add(b);

      return this;
    }

    public TraceBuilder totalNumberOfPaddings(final BigInteger b) {
      if (filled.get(74)) {
        throw new IllegalStateException("TOTAL_NUMBER_OF_PADDINGS already set");
      } else {
        filled.set(74);
      }

      totalNumberOfPaddings.add(b);

      return this;
    }

    public TraceBuilder totalNumberOfReads(final BigInteger b) {
      if (filled.get(75)) {
        throw new IllegalStateException("TOTAL_NUMBER_OF_READS already set");
      } else {
        filled.set(75);
      }

      totalNumberOfReads.add(b);

      return this;
    }

    public TraceBuilder valHi(final BigInteger b) {
      if (filled.get(77)) {
        throw new IllegalStateException("VAL_HI already set");
      } else {
        filled.set(77);
      }

      valHi.add(b);

      return this;
    }

    public TraceBuilder valLo(final BigInteger b) {
      if (filled.get(78)) {
        throw new IllegalStateException("VAL_LO already set");
      } else {
        filled.set(78);
      }

      valLo.add(b);

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
        throw new IllegalStateException("ACC_7 has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("ACC_8 has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("ALIGNED has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("BIT_1 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("BIT_2 has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("BIT_3 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("BIT_4 has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("BIT_5 has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("BIT_6 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BIT_7 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("BIT_8 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("BYTE_3 has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("BYTE_4 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("BYTE_5 has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("BYTE_6 has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("BYTE_7 has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("BYTE_8 has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("CALL_DATA_OFFSET has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("CALL_DATA_SIZE has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("CALL_STACK_DEPTH has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("CALLER has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("CONTEXT_NUMBER has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("CONTEXT_SOURCE has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("CONTEXT_TARGET has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("ERF has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("EXO_IS_HASH has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("EXO_IS_LOG has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("EXO_IS_ROM has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("EXO_IS_TXCD has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("FAST has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("INFO has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("INSTRUCTION has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("IS_DATA has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("IS_MICRO_INSTRUCTION has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("MICRO_INSTRUCTION has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("MICRO_INSTRUCTION_STAMP has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("MIN has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("NIB_1 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("NIB_2 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("NIB_3 has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("NIB_4 has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("NIB_5 has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("NIB_6 has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("NIB_7 has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("NIB_8 has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("NIB_9 has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("OFF_1_LO has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("OFF_2_HI has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("OFF_2_LO has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("OFFSET_OUT_OF_BOUNDS has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("PRECOMPUTATION has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("RAM_STAMP has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException("REFO has not been filled");
      }

      if (!filled.get(62)) {
        throw new IllegalStateException("REFS has not been filled");
      }

      if (!filled.get(64)) {
        throw new IllegalStateException("RETURN_CAPACITY has not been filled");
      }

      if (!filled.get(65)) {
        throw new IllegalStateException("RETURN_OFFSET has not been filled");
      }

      if (!filled.get(63)) {
        throw new IllegalStateException("RETURNER has not been filled");
      }

      if (!filled.get(66)) {
        throw new IllegalStateException("SIZE has not been filled");
      }

      if (!filled.get(67)) {
        throw new IllegalStateException("SIZE_IMPORTED has not been filled");
      }

      if (!filled.get(68)) {
        throw new IllegalStateException("SOURCE_BYTE_OFFSET has not been filled");
      }

      if (!filled.get(69)) {
        throw new IllegalStateException("SOURCE_LIMB_OFFSET has not been filled");
      }

      if (!filled.get(70)) {
        throw new IllegalStateException("TARGET_BYTE_OFFSET has not been filled");
      }

      if (!filled.get(71)) {
        throw new IllegalStateException("TARGET_LIMB_OFFSET has not been filled");
      }

      if (!filled.get(72)) {
        throw new IllegalStateException("TERNARY has not been filled");
      }

      if (!filled.get(76)) {
        throw new IllegalStateException("TO_RAM has not been filled");
      }

      if (!filled.get(73)) {
        throw new IllegalStateException("TOTAL_NUMBER_OF_MICRO_INSTRUCTIONS has not been filled");
      }

      if (!filled.get(74)) {
        throw new IllegalStateException("TOTAL_NUMBER_OF_PADDINGS has not been filled");
      }

      if (!filled.get(75)) {
        throw new IllegalStateException("TOTAL_NUMBER_OF_READS has not been filled");
      }

      if (!filled.get(77)) {
        throw new IllegalStateException("VAL_HI has not been filled");
      }

      if (!filled.get(78)) {
        throw new IllegalStateException("VAL_LO has not been filled");
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
        acc7.add(BigInteger.ZERO);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        acc8.add(BigInteger.ZERO);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        aligned.add(BigInteger.ZERO);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        bit1.add(false);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        bit2.add(false);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        bit3.add(false);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        bit4.add(false);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        bit5.add(false);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        bit6.add(false);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        bit7.add(false);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        bit8.add(false);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        byte2.add(UnsignedByte.of(0));
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        byte3.add(UnsignedByte.of(0));
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        byte4.add(UnsignedByte.of(0));
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        byte5.add(UnsignedByte.of(0));
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        byte6.add(UnsignedByte.of(0));
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        byte7.add(UnsignedByte.of(0));
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        byte8.add(UnsignedByte.of(0));
        this.filled.set(24);
      }
      if (!filled.get(26)) {
        callDataOffset.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        callDataSize.add(BigInteger.ZERO);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        callStackDepth.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(25)) {
        caller.add(BigInteger.ZERO);
        this.filled.set(25);
      }
      if (!filled.get(29)) {
        contextNumber.add(BigInteger.ZERO);
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        contextSource.add(BigInteger.ZERO);
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        contextTarget.add(BigInteger.ZERO);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        counter.add(BigInteger.ZERO);
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        erf.add(false);
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        exoIsHash.add(false);
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        exoIsLog.add(false);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        exoIsRom.add(false);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        exoIsTxcd.add(false);
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        fast.add(BigInteger.ZERO);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        info.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        instruction.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        isData.add(false);
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        isMicroInstruction.add(false);
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        microInstruction.add(BigInteger.ZERO);
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        microInstructionStamp.add(BigInteger.ZERO);
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        min.add(BigInteger.ZERO);
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        nib1.add(UnsignedByte.of(0));
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        nib2.add(UnsignedByte.of(0));
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        nib3.add(UnsignedByte.of(0));
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        nib4.add(UnsignedByte.of(0));
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        nib5.add(UnsignedByte.of(0));
        this.filled.set(50);
      }
      if (!filled.get(51)) {
        nib6.add(UnsignedByte.of(0));
        this.filled.set(51);
      }
      if (!filled.get(52)) {
        nib7.add(UnsignedByte.of(0));
        this.filled.set(52);
      }
      if (!filled.get(53)) {
        nib8.add(UnsignedByte.of(0));
        this.filled.set(53);
      }
      if (!filled.get(54)) {
        nib9.add(UnsignedByte.of(0));
        this.filled.set(54);
      }
      if (!filled.get(56)) {
        off1Lo.add(BigInteger.ZERO);
        this.filled.set(56);
      }
      if (!filled.get(57)) {
        off2Hi.add(BigInteger.ZERO);
        this.filled.set(57);
      }
      if (!filled.get(58)) {
        off2Lo.add(BigInteger.ZERO);
        this.filled.set(58);
      }
      if (!filled.get(55)) {
        offsetOutOfBounds.add(false);
        this.filled.set(55);
      }
      if (!filled.get(59)) {
        precomputation.add(BigInteger.ZERO);
        this.filled.set(59);
      }
      if (!filled.get(60)) {
        ramStamp.add(BigInteger.ZERO);
        this.filled.set(60);
      }
      if (!filled.get(61)) {
        refo.add(BigInteger.ZERO);
        this.filled.set(61);
      }
      if (!filled.get(62)) {
        refs.add(BigInteger.ZERO);
        this.filled.set(62);
      }
      if (!filled.get(64)) {
        returnCapacity.add(BigInteger.ZERO);
        this.filled.set(64);
      }
      if (!filled.get(65)) {
        returnOffset.add(BigInteger.ZERO);
        this.filled.set(65);
      }
      if (!filled.get(63)) {
        returner.add(BigInteger.ZERO);
        this.filled.set(63);
      }
      if (!filled.get(66)) {
        size.add(BigInteger.ZERO);
        this.filled.set(66);
      }
      if (!filled.get(67)) {
        sizeImported.add(BigInteger.ZERO);
        this.filled.set(67);
      }
      if (!filled.get(68)) {
        sourceByteOffset.add(BigInteger.ZERO);
        this.filled.set(68);
      }
      if (!filled.get(69)) {
        sourceLimbOffset.add(BigInteger.ZERO);
        this.filled.set(69);
      }
      if (!filled.get(70)) {
        targetByteOffset.add(BigInteger.ZERO);
        this.filled.set(70);
      }
      if (!filled.get(71)) {
        targetLimbOffset.add(BigInteger.ZERO);
        this.filled.set(71);
      }
      if (!filled.get(72)) {
        ternary.add(BigInteger.ZERO);
        this.filled.set(72);
      }
      if (!filled.get(76)) {
        toRam.add(false);
        this.filled.set(76);
      }
      if (!filled.get(73)) {
        totalNumberOfMicroInstructions.add(BigInteger.ZERO);
        this.filled.set(73);
      }
      if (!filled.get(74)) {
        totalNumberOfPaddings.add(BigInteger.ZERO);
        this.filled.set(74);
      }
      if (!filled.get(75)) {
        totalNumberOfReads.add(BigInteger.ZERO);
        this.filled.set(75);
      }
      if (!filled.get(77)) {
        valHi.add(BigInteger.ZERO);
        this.filled.set(77);
      }
      if (!filled.get(78)) {
        valLo.add(BigInteger.ZERO);
        this.filled.set(78);
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
          acc7,
          acc8,
          aligned,
          bit1,
          bit2,
          bit3,
          bit4,
          bit5,
          bit6,
          bit7,
          bit8,
          byte1,
          byte2,
          byte3,
          byte4,
          byte5,
          byte6,
          byte7,
          byte8,
          callDataOffset,
          callDataSize,
          callStackDepth,
          caller,
          contextNumber,
          contextSource,
          contextTarget,
          counter,
          erf,
          exoIsHash,
          exoIsLog,
          exoIsRom,
          exoIsTxcd,
          fast,
          info,
          instruction,
          isData,
          isMicroInstruction,
          microInstruction,
          microInstructionStamp,
          min,
          nib1,
          nib2,
          nib3,
          nib4,
          nib5,
          nib6,
          nib7,
          nib8,
          nib9,
          off1Lo,
          off2Hi,
          off2Lo,
          offsetOutOfBounds,
          precomputation,
          ramStamp,
          refo,
          refs,
          returnCapacity,
          returnOffset,
          returner,
          size,
          sizeImported,
          sourceByteOffset,
          sourceLimbOffset,
          targetByteOffset,
          targetLimbOffset,
          ternary,
          toRam,
          totalNumberOfMicroInstructions,
          totalNumberOfPaddings,
          totalNumberOfReads,
          valHi,
          valLo);
    }
  }
}
