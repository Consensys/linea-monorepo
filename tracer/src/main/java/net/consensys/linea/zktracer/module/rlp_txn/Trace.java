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

package net.consensys.linea.zktracer.module.rlp_txn;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.List;

import com.fasterxml.jackson.annotation.JsonProperty;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically. Any modifications to this code may be overwritten
 * and could lead to unexpected behavior. Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
record Trace(
    @JsonProperty("ABS_TX_NUM") List<BigInteger> absTxNum,
    @JsonProperty("ABS_TX_NUM_INFINY") List<BigInteger> absTxNumInfiny,
    @JsonProperty("ACC_1") List<BigInteger> acc1,
    @JsonProperty("ACC_2") List<BigInteger> acc2,
    @JsonProperty("ACC_BYTESIZE") List<BigInteger> accBytesize,
    @JsonProperty("ACCESS_TUPLE_BYTESIZE") List<BigInteger> accessTupleBytesize,
    @JsonProperty("ADDR_HI") List<BigInteger> addrHi,
    @JsonProperty("ADDR_LO") List<BigInteger> addrLo,
    @JsonProperty("BIT") List<Boolean> bit,
    @JsonProperty("BIT_ACC") List<UnsignedByte> bitAcc,
    @JsonProperty("BYTE_1") List<UnsignedByte> byte1,
    @JsonProperty("BYTE_2") List<UnsignedByte> byte2,
    @JsonProperty("CODE_FRAGMENT_INDEX") List<BigInteger> codeFragmentIndex,
    @JsonProperty("COUNTER") List<UnsignedByte> counter,
    @JsonProperty("DATA_HI") List<BigInteger> dataHi,
    @JsonProperty("DATA_LO") List<BigInteger> dataLo,
    @JsonProperty("DATAGASCOST") List<BigInteger> datagascost,
    @JsonProperty("DEPTH_1") List<Boolean> depth1,
    @JsonProperty("DEPTH_2") List<Boolean> depth2,
    @JsonProperty("DONE") List<Boolean> done,
    @JsonProperty("end_phase") List<Boolean> endPhase,
    @JsonProperty("INDEX_DATA") List<BigInteger> indexData,
    @JsonProperty("INDEX_LT") List<BigInteger> indexLt,
    @JsonProperty("INDEX_LX") List<BigInteger> indexLx,
    @JsonProperty("INPUT_1") List<BigInteger> input1,
    @JsonProperty("INPUT_2") List<BigInteger> input2,
    @JsonProperty("is_padding") List<Boolean> isPadding,
    @JsonProperty("is_prefix") List<Boolean> isPrefix,
    @JsonProperty("LIMB") List<BigInteger> limb,
    @JsonProperty("LIMB_CONSTRUCTED") List<Boolean> limbConstructed,
    @JsonProperty("LT") List<Boolean> lt,
    @JsonProperty("LX") List<Boolean> lx,
    @JsonProperty("nBYTES") List<UnsignedByte> nBytes,
    @JsonProperty("nb_Addr") List<BigInteger> nbAddr,
    @JsonProperty("nb_Sto") List<BigInteger> nbSto,
    @JsonProperty("nb_Sto_per_Addr") List<BigInteger> nbStoPerAddr,
    @JsonProperty("number_step") List<UnsignedByte> numberStep,
    @JsonProperty("PHASE_0") List<Boolean> phase0,
    @JsonProperty("PHASE_1") List<Boolean> phase1,
    @JsonProperty("PHASE_10") List<Boolean> phase10,
    @JsonProperty("PHASE_11") List<Boolean> phase11,
    @JsonProperty("PHASE_12") List<Boolean> phase12,
    @JsonProperty("PHASE_13") List<Boolean> phase13,
    @JsonProperty("PHASE_14") List<Boolean> phase14,
    @JsonProperty("PHASE_2") List<Boolean> phase2,
    @JsonProperty("PHASE_3") List<Boolean> phase3,
    @JsonProperty("PHASE_4") List<Boolean> phase4,
    @JsonProperty("PHASE_5") List<Boolean> phase5,
    @JsonProperty("PHASE_6") List<Boolean> phase6,
    @JsonProperty("PHASE_7") List<Boolean> phase7,
    @JsonProperty("PHASE_8") List<Boolean> phase8,
    @JsonProperty("PHASE_9") List<Boolean> phase9,
    @JsonProperty("PHASE_BYTESIZE") List<BigInteger> phaseBytesize,
    @JsonProperty("POWER") List<BigInteger> power,
    @JsonProperty("REQUIRES_EVM_EXECUTION") List<Boolean> requiresEvmExecution,
    @JsonProperty("RLP_LT_BYTESIZE") List<BigInteger> rlpLtBytesize,
    @JsonProperty("RLP_LX_BYTESIZE") List<BigInteger> rlpLxBytesize,
    @JsonProperty("TYPE") List<UnsignedByte> type) {
  static TraceBuilder builder() {
    return new TraceBuilder();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ABS_TX_NUM")
    private final List<BigInteger> absTxNum = new ArrayList<>();

    @JsonProperty("ABS_TX_NUM_INFINY")
    private final List<BigInteger> absTxNumInfiny = new ArrayList<>();

    @JsonProperty("ACC_1")
    private final List<BigInteger> acc1 = new ArrayList<>();

    @JsonProperty("ACC_2")
    private final List<BigInteger> acc2 = new ArrayList<>();

    @JsonProperty("ACC_BYTESIZE")
    private final List<BigInteger> accBytesize = new ArrayList<>();

    @JsonProperty("ACCESS_TUPLE_BYTESIZE")
    private final List<BigInteger> accessTupleBytesize = new ArrayList<>();

    @JsonProperty("ADDR_HI")
    private final List<BigInteger> addrHi = new ArrayList<>();

    @JsonProperty("ADDR_LO")
    private final List<BigInteger> addrLo = new ArrayList<>();

    @JsonProperty("BIT")
    private final List<Boolean> bit = new ArrayList<>();

    @JsonProperty("BIT_ACC")
    private final List<UnsignedByte> bitAcc = new ArrayList<>();

    @JsonProperty("BYTE_1")
    private final List<UnsignedByte> byte1 = new ArrayList<>();

    @JsonProperty("BYTE_2")
    private final List<UnsignedByte> byte2 = new ArrayList<>();

    @JsonProperty("CODE_FRAGMENT_INDEX")
    private final List<BigInteger> codeFragmentIndex = new ArrayList<>();

    @JsonProperty("COUNTER")
    private final List<UnsignedByte> counter = new ArrayList<>();

    @JsonProperty("DATA_HI")
    private final List<BigInteger> dataHi = new ArrayList<>();

    @JsonProperty("DATA_LO")
    private final List<BigInteger> dataLo = new ArrayList<>();

    @JsonProperty("DATAGASCOST")
    private final List<BigInteger> datagascost = new ArrayList<>();

    @JsonProperty("DEPTH_1")
    private final List<Boolean> depth1 = new ArrayList<>();

    @JsonProperty("DEPTH_2")
    private final List<Boolean> depth2 = new ArrayList<>();

    @JsonProperty("DONE")
    private final List<Boolean> done = new ArrayList<>();

    @JsonProperty("end_phase")
    private final List<Boolean> endPhase = new ArrayList<>();

    @JsonProperty("INDEX_DATA")
    private final List<BigInteger> indexData = new ArrayList<>();

    @JsonProperty("INDEX_LT")
    private final List<BigInteger> indexLt = new ArrayList<>();

    @JsonProperty("INDEX_LX")
    private final List<BigInteger> indexLx = new ArrayList<>();

    @JsonProperty("INPUT_1")
    private final List<BigInteger> input1 = new ArrayList<>();

    @JsonProperty("INPUT_2")
    private final List<BigInteger> input2 = new ArrayList<>();

    @JsonProperty("is_padding")
    private final List<Boolean> isPadding = new ArrayList<>();

    @JsonProperty("is_prefix")
    private final List<Boolean> isPrefix = new ArrayList<>();

    @JsonProperty("LIMB")
    private final List<BigInteger> limb = new ArrayList<>();

    @JsonProperty("LIMB_CONSTRUCTED")
    private final List<Boolean> limbConstructed = new ArrayList<>();

    @JsonProperty("LT")
    private final List<Boolean> lt = new ArrayList<>();

    @JsonProperty("LX")
    private final List<Boolean> lx = new ArrayList<>();

    @JsonProperty("nBYTES")
    private final List<UnsignedByte> nBytes = new ArrayList<>();

    @JsonProperty("nb_Addr")
    private final List<BigInteger> nbAddr = new ArrayList<>();

    @JsonProperty("nb_Sto")
    private final List<BigInteger> nbSto = new ArrayList<>();

    @JsonProperty("nb_Sto_per_Addr")
    private final List<BigInteger> nbStoPerAddr = new ArrayList<>();

    @JsonProperty("number_step")
    private final List<UnsignedByte> numberStep = new ArrayList<>();

    @JsonProperty("PHASE_0")
    private final List<Boolean> phase0 = new ArrayList<>();

    @JsonProperty("PHASE_1")
    private final List<Boolean> phase1 = new ArrayList<>();

    @JsonProperty("PHASE_10")
    private final List<Boolean> phase10 = new ArrayList<>();

    @JsonProperty("PHASE_11")
    private final List<Boolean> phase11 = new ArrayList<>();

    @JsonProperty("PHASE_12")
    private final List<Boolean> phase12 = new ArrayList<>();

    @JsonProperty("PHASE_13")
    private final List<Boolean> phase13 = new ArrayList<>();

    @JsonProperty("PHASE_14")
    private final List<Boolean> phase14 = new ArrayList<>();

    @JsonProperty("PHASE_2")
    private final List<Boolean> phase2 = new ArrayList<>();

    @JsonProperty("PHASE_3")
    private final List<Boolean> phase3 = new ArrayList<>();

    @JsonProperty("PHASE_4")
    private final List<Boolean> phase4 = new ArrayList<>();

    @JsonProperty("PHASE_5")
    private final List<Boolean> phase5 = new ArrayList<>();

    @JsonProperty("PHASE_6")
    private final List<Boolean> phase6 = new ArrayList<>();

    @JsonProperty("PHASE_7")
    private final List<Boolean> phase7 = new ArrayList<>();

    @JsonProperty("PHASE_8")
    private final List<Boolean> phase8 = new ArrayList<>();

    @JsonProperty("PHASE_9")
    private final List<Boolean> phase9 = new ArrayList<>();

    @JsonProperty("PHASE_BYTESIZE")
    private final List<BigInteger> phaseBytesize = new ArrayList<>();

    @JsonProperty("POWER")
    private final List<BigInteger> power = new ArrayList<>();

    @JsonProperty("REQUIRES_EVM_EXECUTION")
    private final List<Boolean> requiresEvmExecution = new ArrayList<>();

    @JsonProperty("RLP_LT_BYTESIZE")
    private final List<BigInteger> rlpLtBytesize = new ArrayList<>();

    @JsonProperty("RLP_LX_BYTESIZE")
    private final List<BigInteger> rlpLxBytesize = new ArrayList<>();

    @JsonProperty("TYPE")
    private final List<UnsignedByte> type = new ArrayList<>();

    private TraceBuilder() {}

    int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.absTxNum.size();
    }

    TraceBuilder absTxNum(final BigInteger b) {
      if (filled.get(20)) {
        throw new IllegalStateException("ABS_TX_NUM already set");
      } else {
        filled.set(20);
      }

      absTxNum.add(b);

      return this;
    }

    TraceBuilder absTxNumInfiny(final BigInteger b) {
      if (filled.get(28)) {
        throw new IllegalStateException("ABS_TX_NUM_INFINY already set");
      } else {
        filled.set(28);
      }

      absTxNumInfiny.add(b);

      return this;
    }

    TraceBuilder acc1(final Bytes b) {
      if (filled.get(56)) {
        throw new IllegalStateException("ACC_1 already set");
      } else {
        filled.set(56);
      }

      acc1.add(b.toUnsignedBigInteger());

      return this;
    }

    TraceBuilder acc2(final Bytes b) {
      if (filled.get(32)) {
        throw new IllegalStateException("ACC_2 already set");
      } else {
        filled.set(32);
      }

      acc2.add(b.toUnsignedBigInteger());

      return this;
    }

    TraceBuilder accBytesize(final BigInteger b) {
      if (filled.get(10)) {
        throw new IllegalStateException("ACC_BYTESIZE already set");
      } else {
        filled.set(10);
      }

      accBytesize.add(b);

      return this;
    }

    TraceBuilder accessTupleBytesize(final BigInteger b) {
      if (filled.get(55)) {
        throw new IllegalStateException("ACCESS_TUPLE_BYTESIZE already set");
      } else {
        filled.set(55);
      }

      accessTupleBytesize.add(b);

      return this;
    }

    TraceBuilder addrHi(final Bytes b) {
      if (filled.get(46)) {
        throw new IllegalStateException("ADDR_HI already set");
      } else {
        filled.set(46);
      }

      addrHi.add(b.toUnsignedBigInteger());

      return this;
    }

    TraceBuilder addrLo(final Bytes b) {
      if (filled.get(22)) {
        throw new IllegalStateException("ADDR_LO already set");
      } else {
        filled.set(22);
      }

      addrLo.add(b.toUnsignedBigInteger());

      return this;
    }

    TraceBuilder bit(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("BIT already set");
      } else {
        filled.set(13);
      }

      bit.add(b);

      return this;
    }

    TraceBuilder bitAcc(final UnsignedByte b) {
      if (filled.get(8)) {
        throw new IllegalStateException("BIT_ACC already set");
      } else {
        filled.set(8);
      }

      bitAcc.add(b);

      return this;
    }

    TraceBuilder byte1(final UnsignedByte b) {
      if (filled.get(25)) {
        throw new IllegalStateException("BYTE_1 already set");
      } else {
        filled.set(25);
      }

      byte1.add(b);

      return this;
    }

    TraceBuilder byte2(final UnsignedByte b) {
      if (filled.get(15)) {
        throw new IllegalStateException("BYTE_2 already set");
      } else {
        filled.set(15);
      }

      byte2.add(b);

      return this;
    }

    TraceBuilder codeFragmentIndex(final BigInteger b) {
      if (filled.get(36)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX already set");
      } else {
        filled.set(36);
      }

      codeFragmentIndex.add(b);

      return this;
    }

    TraceBuilder counter(final BigInteger b) {
      if (filled.get(35)) {
        throw new IllegalStateException("COUNTER already set");
      } else {
        filled.set(35);
      }

      counter.add(UnsignedByte.of(b.longValueExact()));

      return this;
    }

    TraceBuilder dataHi(final BigInteger b) {
      if (filled.get(40)) {
        throw new IllegalStateException("DATA_HI already set");
      } else {
        filled.set(40);
      }

      dataHi.add(b);

      return this;
    }

    TraceBuilder dataLo(final BigInteger b) {
      if (filled.get(11)) {
        throw new IllegalStateException("DATA_LO already set");
      } else {
        filled.set(11);
      }

      dataLo.add(b);

      return this;
    }

    TraceBuilder datagascost(final BigInteger b) {
      if (filled.get(33)) {
        throw new IllegalStateException("DATAGASCOST already set");
      } else {
        filled.set(33);
      }

      datagascost.add(b);

      return this;
    }

    TraceBuilder depth1(final Boolean b) {
      if (filled.get(7)) {
        throw new IllegalStateException("DEPTH_1 already set");
      } else {
        filled.set(7);
      }

      depth1.add(b);

      return this;
    }

    TraceBuilder depth2(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("DEPTH_2 already set");
      } else {
        filled.set(58);
      }

      depth2.add(b);

      return this;
    }

    TraceBuilder done(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("DONE already set");
      } else {
        filled.set(31);
      }

      done.add(b);

      return this;
    }

    TraceBuilder endPhase(final Boolean b) {
      if (filled.get(5)) {
        throw new IllegalStateException("end_phase already set");
      } else {
        filled.set(5);
      }

      endPhase.add(b);

      return this;
    }

    TraceBuilder indexData(final BigInteger b) {
      if (filled.get(39)) {
        throw new IllegalStateException("INDEX_DATA already set");
      } else {
        filled.set(39);
      }

      indexData.add(b);

      return this;
    }

    TraceBuilder indexLt(final BigInteger b) {
      if (filled.get(51)) {
        throw new IllegalStateException("INDEX_LT already set");
      } else {
        filled.set(51);
      }

      indexLt.add(b);

      return this;
    }

    TraceBuilder indexLx(final BigInteger b) {
      if (filled.get(42)) {
        throw new IllegalStateException("INDEX_LX already set");
      } else {
        filled.set(42);
      }

      indexLx.add(b);

      return this;
    }

    TraceBuilder input1(final Bytes b) {
      if (filled.get(49)) {
        throw new IllegalStateException("INPUT_1 already set");
      } else {
        filled.set(49);
      }

      input1.add(b.toUnsignedBigInteger());

      return this;
    }

    TraceBuilder input2(final Bytes b) {
      if (filled.get(17)) {
        throw new IllegalStateException("INPUT_2 already set");
      } else {
        filled.set(17);
      }

      input2.add(b.toUnsignedBigInteger());

      return this;
    }

    TraceBuilder isPadding(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("is_padding already set");
      } else {
        filled.set(54);
      }

      isPadding.add(b);

      return this;
    }

    TraceBuilder isPrefix(final Boolean b) {
      if (filled.get(38)) {
        throw new IllegalStateException("is_prefix already set");
      } else {
        filled.set(38);
      }

      isPrefix.add(b);

      return this;
    }

    TraceBuilder limb(final BigInteger b) {
      if (filled.get(4)) {
        throw new IllegalStateException("LIMB already set");
      } else {
        filled.set(4);
      }

      limb.add(b);

      return this;
    }

    TraceBuilder limbConstructed(final Boolean b) {
      if (filled.get(27)) {
        throw new IllegalStateException("LIMB_CONSTRUCTED already set");
      } else {
        filled.set(27);
      }

      limbConstructed.add(b);

      return this;
    }

    TraceBuilder lt(final Boolean b) {
      if (filled.get(23)) {
        throw new IllegalStateException("LT already set");
      } else {
        filled.set(23);
      }

      lt.add(b);

      return this;
    }

    TraceBuilder lx(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("LX already set");
      } else {
        filled.set(24);
      }

      lx.add(b);

      return this;
    }

    TraceBuilder nBytes(final UnsignedByte b) {
      if (filled.get(1)) {
        throw new IllegalStateException("nBYTES already set");
      } else {
        filled.set(1);
      }

      nBytes.add(b);

      return this;
    }

    TraceBuilder nbAddr(final BigInteger b) {
      if (filled.get(50)) {
        throw new IllegalStateException("nb_Addr already set");
      } else {
        filled.set(50);
      }

      nbAddr.add(b);

      return this;
    }

    TraceBuilder nbSto(final BigInteger b) {
      if (filled.get(37)) {
        throw new IllegalStateException("nb_Sto already set");
      } else {
        filled.set(37);
      }

      nbSto.add(b);

      return this;
    }

    TraceBuilder nbStoPerAddr(final BigInteger b) {
      if (filled.get(26)) {
        throw new IllegalStateException("nb_Sto_per_Addr already set");
      } else {
        filled.set(26);
      }

      nbStoPerAddr.add(b);

      return this;
    }

    TraceBuilder numberStep(final UnsignedByte b) {
      if (filled.get(6)) {
        throw new IllegalStateException("number_step already set");
      } else {
        filled.set(6);
      }

      numberStep.add(b);

      return this;
    }

    TraceBuilder phase0(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("PHASE_0 already set");
      } else {
        filled.set(34);
      }

      phase0.add(b);

      return this;
    }

    TraceBuilder phase1(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("PHASE_1 already set");
      } else {
        filled.set(18);
      }

      phase1.add(b);

      return this;
    }

    TraceBuilder phase10(final Boolean b) {
      if (filled.get(30)) {
        throw new IllegalStateException("PHASE_10 already set");
      } else {
        filled.set(30);
      }

      phase10.add(b);

      return this;
    }

    TraceBuilder phase11(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("PHASE_11 already set");
      } else {
        filled.set(52);
      }

      phase11.add(b);

      return this;
    }

    TraceBuilder phase12(final Boolean b) {
      if (filled.get(47)) {
        throw new IllegalStateException("PHASE_12 already set");
      } else {
        filled.set(47);
      }

      phase12.add(b);

      return this;
    }

    TraceBuilder phase13(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("PHASE_13 already set");
      } else {
        filled.set(48);
      }

      phase13.add(b);

      return this;
    }

    TraceBuilder phase14(final Boolean b) {
      if (filled.get(60)) {
        throw new IllegalStateException("PHASE_14 already set");
      } else {
        filled.set(60);
      }

      phase14.add(b);

      return this;
    }

    TraceBuilder phase2(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("PHASE_2 already set");
      } else {
        filled.set(9);
      }

      phase2.add(b);

      return this;
    }

    TraceBuilder phase3(final Boolean b) {
      if (filled.get(2)) {
        throw new IllegalStateException("PHASE_3 already set");
      } else {
        filled.set(2);
      }

      phase3.add(b);

      return this;
    }

    TraceBuilder phase4(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("PHASE_4 already set");
      } else {
        filled.set(12);
      }

      phase4.add(b);

      return this;
    }

    TraceBuilder phase5(final Boolean b) {
      if (filled.get(3)) {
        throw new IllegalStateException("PHASE_5 already set");
      } else {
        filled.set(3);
      }

      phase5.add(b);

      return this;
    }

    TraceBuilder phase6(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("PHASE_6 already set");
      } else {
        filled.set(41);
      }

      phase6.add(b);

      return this;
    }

    TraceBuilder phase7(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("PHASE_7 already set");
      } else {
        filled.set(57);
      }

      phase7.add(b);

      return this;
    }

    TraceBuilder phase8(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("PHASE_8 already set");
      } else {
        filled.set(19);
      }

      phase8.add(b);

      return this;
    }

    TraceBuilder phase9(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("PHASE_9 already set");
      } else {
        filled.set(16);
      }

      phase9.add(b);

      return this;
    }

    TraceBuilder phaseBytesize(final BigInteger b) {
      if (filled.get(21)) {
        throw new IllegalStateException("PHASE_BYTESIZE already set");
      } else {
        filled.set(21);
      }

      phaseBytesize.add(b);

      return this;
    }

    TraceBuilder power(final BigInteger b) {
      if (filled.get(53)) {
        throw new IllegalStateException("POWER already set");
      } else {
        filled.set(53);
      }

      power.add(b);

      return this;
    }

    TraceBuilder requiresEvmExecution(final Boolean b) {
      if (filled.get(0)) {
        throw new IllegalStateException("REQUIRES_EVM_EXECUTION already set");
      } else {
        filled.set(0);
      }

      requiresEvmExecution.add(b);

      return this;
    }

    TraceBuilder rlpLtBytesize(final BigInteger b) {
      if (filled.get(59)) {
        throw new IllegalStateException("RLP_LT_BYTESIZE already set");
      } else {
        filled.set(59);
      }

      rlpLtBytesize.add(b);

      return this;
    }

    TraceBuilder rlpLxBytesize(final BigInteger b) {
      if (filled.get(45)) {
        throw new IllegalStateException("RLP_LX_BYTESIZE already set");
      } else {
        filled.set(45);
      }

      rlpLxBytesize.add(b);

      return this;
    }

    TraceBuilder type(final UnsignedByte b) {
      if (filled.get(43)) {
        throw new IllegalStateException("TYPE already set");
      } else {
        filled.set(43);
      }

      type.add(b);

      return this;
    }

    TraceBuilder setAbsTxNumAt(final BigInteger b, int i) {
      absTxNum.set(i, b);

      return this;
    }

    TraceBuilder setAbsTxNumInfinyAt(final BigInteger b, int i) {
      absTxNumInfiny.set(i, b);

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

    TraceBuilder setAccBytesizeAt(final BigInteger b, int i) {
      accBytesize.set(i, b);

      return this;
    }

    TraceBuilder setAccessTupleBytesizeAt(final BigInteger b, int i) {
      accessTupleBytesize.set(i, b);

      return this;
    }

    TraceBuilder setAddrHiAt(final BigInteger b, int i) {
      addrHi.set(i, b);

      return this;
    }

    TraceBuilder setAddrLoAt(final BigInteger b, int i) {
      addrLo.set(i, b);

      return this;
    }

    TraceBuilder setBitAt(final Boolean b, int i) {
      bit.set(i, b);

      return this;
    }

    TraceBuilder setBitAccAt(final UnsignedByte b, int i) {
      bitAcc.set(i, b);

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

    TraceBuilder setCodeFragmentIndexAt(final BigInteger b, int i) {
      codeFragmentIndex.set(i, b);

      return this;
    }

    TraceBuilder setCounterAt(final UnsignedByte b, int i) {
      counter.set(i, b);

      return this;
    }

    TraceBuilder setDataHiAt(final BigInteger b, int i) {
      dataHi.set(i, b);

      return this;
    }

    TraceBuilder setDataLoAt(final BigInteger b, int i) {
      dataLo.set(i, b);

      return this;
    }

    TraceBuilder setDatagascostAt(final BigInteger b, int i) {
      datagascost.set(i, b);

      return this;
    }

    TraceBuilder setDepth1At(final Boolean b, int i) {
      depth1.set(i, b);

      return this;
    }

    TraceBuilder setDepth2At(final Boolean b, int i) {
      depth2.set(i, b);

      return this;
    }

    TraceBuilder setDoneAt(final Boolean b, int i) {
      done.set(i, b);

      return this;
    }

    TraceBuilder setEndPhaseAt(final Boolean b, int i) {
      endPhase.set(i, b);

      return this;
    }

    TraceBuilder setIndexDataAt(final BigInteger b, int i) {
      indexData.set(i, b);

      return this;
    }

    TraceBuilder setIndexLtAt(final BigInteger b, int i) {
      indexLt.set(i, b);

      return this;
    }

    TraceBuilder setIndexLxAt(final BigInteger b, int i) {
      indexLx.set(i, b);

      return this;
    }

    TraceBuilder setInput1At(final BigInteger b, int i) {
      input1.set(i, b);

      return this;
    }

    TraceBuilder setInput2At(final BigInteger b, int i) {
      input2.set(i, b);

      return this;
    }

    TraceBuilder setIsPaddingAt(final Boolean b, int i) {
      isPadding.set(i, b);

      return this;
    }

    TraceBuilder setIsPrefixAt(final Boolean b, int i) {
      isPrefix.set(i, b);

      return this;
    }

    TraceBuilder setLimbAt(final BigInteger b, int i) {
      limb.set(i, b);

      return this;
    }

    TraceBuilder setLimbConstructedAt(final Boolean b, int i) {
      limbConstructed.set(i, b);

      return this;
    }

    TraceBuilder setLtAt(final Boolean b, int i) {
      lt.set(i, b);

      return this;
    }

    TraceBuilder setLxAt(final Boolean b, int i) {
      lx.set(i, b);

      return this;
    }

    TraceBuilder setNBytesAt(final UnsignedByte b, int i) {
      nBytes.set(i, b);

      return this;
    }

    TraceBuilder setNbAddrAt(final BigInteger b, int i) {
      nbAddr.set(i, b);

      return this;
    }

    TraceBuilder setNbStoAt(final BigInteger b, int i) {
      nbSto.set(i, b);

      return this;
    }

    TraceBuilder setNbStoPerAddrAt(final BigInteger b, int i) {
      nbStoPerAddr.set(i, b);

      return this;
    }

    TraceBuilder setNumberStepAt(final UnsignedByte b, int i) {
      numberStep.set(i, b);

      return this;
    }

    TraceBuilder setPhase0At(final Boolean b, int i) {
      phase0.set(i, b);

      return this;
    }

    TraceBuilder setPhase1At(final Boolean b, int i) {
      phase1.set(i, b);

      return this;
    }

    TraceBuilder setPhase10At(final Boolean b, int i) {
      phase10.set(i, b);

      return this;
    }

    TraceBuilder setPhase11At(final Boolean b, int i) {
      phase11.set(i, b);

      return this;
    }

    TraceBuilder setPhase12At(final Boolean b, int i) {
      phase12.set(i, b);

      return this;
    }

    TraceBuilder setPhase13At(final Boolean b, int i) {
      phase13.set(i, b);

      return this;
    }

    TraceBuilder setPhase14At(final Boolean b, int i) {
      phase14.set(i, b);

      return this;
    }

    TraceBuilder setPhase2At(final Boolean b, int i) {
      phase2.set(i, b);

      return this;
    }

    TraceBuilder setPhase3At(final Boolean b, int i) {
      phase3.set(i, b);

      return this;
    }

    TraceBuilder setPhase4At(final Boolean b, int i) {
      phase4.set(i, b);

      return this;
    }

    TraceBuilder setPhase5At(final Boolean b, int i) {
      phase5.set(i, b);

      return this;
    }

    TraceBuilder setPhase6At(final Boolean b, int i) {
      phase6.set(i, b);

      return this;
    }

    TraceBuilder setPhase7At(final Boolean b, int i) {
      phase7.set(i, b);

      return this;
    }

    TraceBuilder setPhase8At(final Boolean b, int i) {
      phase8.set(i, b);

      return this;
    }

    TraceBuilder setPhase9At(final Boolean b, int i) {
      phase9.set(i, b);

      return this;
    }

    TraceBuilder setPhaseBytesizeAt(final BigInteger b, int i) {
      phaseBytesize.set(i, b);

      return this;
    }

    TraceBuilder setPowerAt(final BigInteger b, int i) {
      power.set(i, b);

      return this;
    }

    TraceBuilder setRequiresEvmExecutionAt(final Boolean b, int i) {
      requiresEvmExecution.set(i, b);

      return this;
    }

    TraceBuilder setRlpLtBytesizeAt(final BigInteger b, int i) {
      rlpLtBytesize.set(i, b);

      return this;
    }

    TraceBuilder setRlpLxBytesizeAt(final BigInteger b, int i) {
      rlpLxBytesize.set(i, b);

      return this;
    }

    TraceBuilder setTypeAt(final UnsignedByte b, int i) {
      type.set(i, b);

      return this;
    }

    TraceBuilder setAbsTxNumRelative(final BigInteger b, int i) {
      absTxNum.set(absTxNum.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAbsTxNumInfinyRelative(final BigInteger b, int i) {
      absTxNumInfiny.set(absTxNumInfiny.size() - 1 - i, b);

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

    TraceBuilder setAccBytesizeRelative(final BigInteger b, int i) {
      accBytesize.set(accBytesize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAccessTupleBytesizeRelative(final BigInteger b, int i) {
      accessTupleBytesize.set(accessTupleBytesize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAddrHiRelative(final BigInteger b, int i) {
      addrHi.set(addrHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setAddrLoRelative(final BigInteger b, int i) {
      addrLo.set(addrLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitRelative(final Boolean b, int i) {
      bit.set(bit.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setBitAccRelative(final UnsignedByte b, int i) {
      bitAcc.set(bitAcc.size() - 1 - i, b);

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

    TraceBuilder setCodeFragmentIndexRelative(final BigInteger b, int i) {
      codeFragmentIndex.set(codeFragmentIndex.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setCounterRelative(final UnsignedByte b, int i) {
      counter.set(counter.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDataHiRelative(final BigInteger b, int i) {
      dataHi.set(dataHi.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDataLoRelative(final BigInteger b, int i) {
      dataLo.set(dataLo.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDatagascostRelative(final BigInteger b, int i) {
      datagascost.set(datagascost.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDepth1Relative(final Boolean b, int i) {
      depth1.set(depth1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDepth2Relative(final Boolean b, int i) {
      depth2.set(depth2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setDoneRelative(final Boolean b, int i) {
      done.set(done.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setEndPhaseRelative(final Boolean b, int i) {
      endPhase.set(endPhase.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIndexDataRelative(final BigInteger b, int i) {
      indexData.set(indexData.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIndexLtRelative(final BigInteger b, int i) {
      indexLt.set(indexLt.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIndexLxRelative(final BigInteger b, int i) {
      indexLx.set(indexLx.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInput1Relative(final BigInteger b, int i) {
      input1.set(input1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setInput2Relative(final BigInteger b, int i) {
      input2.set(input2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsPaddingRelative(final Boolean b, int i) {
      isPadding.set(isPadding.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setIsPrefixRelative(final Boolean b, int i) {
      isPrefix.set(isPrefix.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLimbRelative(final BigInteger b, int i) {
      limb.set(limb.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLimbConstructedRelative(final Boolean b, int i) {
      limbConstructed.set(limbConstructed.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLtRelative(final Boolean b, int i) {
      lt.set(lt.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setLxRelative(final Boolean b, int i) {
      lx.set(lx.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNBytesRelative(final UnsignedByte b, int i) {
      nBytes.set(nBytes.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNbAddrRelative(final BigInteger b, int i) {
      nbAddr.set(nbAddr.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNbStoRelative(final BigInteger b, int i) {
      nbSto.set(nbSto.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNbStoPerAddrRelative(final BigInteger b, int i) {
      nbStoPerAddr.set(nbStoPerAddr.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setNumberStepRelative(final UnsignedByte b, int i) {
      numberStep.set(numberStep.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase0Relative(final Boolean b, int i) {
      phase0.set(phase0.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase1Relative(final Boolean b, int i) {
      phase1.set(phase1.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase10Relative(final Boolean b, int i) {
      phase10.set(phase10.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase11Relative(final Boolean b, int i) {
      phase11.set(phase11.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase12Relative(final Boolean b, int i) {
      phase12.set(phase12.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase13Relative(final Boolean b, int i) {
      phase13.set(phase13.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase14Relative(final Boolean b, int i) {
      phase14.set(phase14.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase2Relative(final Boolean b, int i) {
      phase2.set(phase2.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase3Relative(final Boolean b, int i) {
      phase3.set(phase3.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase4Relative(final Boolean b, int i) {
      phase4.set(phase4.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase5Relative(final Boolean b, int i) {
      phase5.set(phase5.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase6Relative(final Boolean b, int i) {
      phase6.set(phase6.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase7Relative(final Boolean b, int i) {
      phase7.set(phase7.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase8Relative(final Boolean b, int i) {
      phase8.set(phase8.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhase9Relative(final Boolean b, int i) {
      phase9.set(phase9.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPhaseBytesizeRelative(final BigInteger b, int i) {
      phaseBytesize.set(phaseBytesize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setPowerRelative(final BigInteger b, int i) {
      power.set(power.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setRequiresEvmExecutionRelative(final Boolean b, int i) {
      requiresEvmExecution.set(requiresEvmExecution.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setRlpLtBytesizeRelative(final BigInteger b, int i) {
      rlpLtBytesize.set(rlpLtBytesize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setRlpLxBytesizeRelative(final BigInteger b, int i) {
      rlpLxBytesize.set(rlpLxBytesize.size() - 1 - i, b);

      return this;
    }

    TraceBuilder setTypeRelative(final UnsignedByte b, int i) {
      type.set(type.size() - 1 - i, b);

      return this;
    }

    TraceBuilder validateRow() {
      if (!filled.get(20)) {
        throw new IllegalStateException("ABS_TX_NUM has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("ABS_TX_NUM_INFINY has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("ACC_1 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("ACC_2 has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("ACC_BYTESIZE has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("ACCESS_TUPLE_BYTESIZE has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("ADDR_HI has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("ADDR_LO has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("BIT has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("BIT_ACC has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("BYTE_1 has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("BYTE_2 has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("CODE_FRAGMENT_INDEX has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("COUNTER has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("DATA_HI has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("DATA_LO has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("DATAGASCOST has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("DEPTH_1 has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("DEPTH_2 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("DONE has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("end_phase has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("INDEX_DATA has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("INDEX_LT has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("INDEX_LX has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("INPUT_1 has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("INPUT_2 has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("is_padding has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("is_prefix has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("LIMB has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("LIMB_CONSTRUCTED has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("LT has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("LX has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("nBYTES has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("nb_Addr has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("nb_Sto has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("nb_Sto_per_Addr has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("number_step has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("PHASE_0 has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("PHASE_1 has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("PHASE_10 has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("PHASE_11 has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("PHASE_12 has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("PHASE_13 has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("PHASE_14 has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("PHASE_2 has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("PHASE_3 has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("PHASE_4 has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("PHASE_5 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("PHASE_6 has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("PHASE_7 has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("PHASE_8 has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("PHASE_9 has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("PHASE_BYTESIZE has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("POWER has not been filled");
      }

      if (!filled.get(0)) {
        throw new IllegalStateException("REQUIRES_EVM_EXECUTION has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("RLP_LT_BYTESIZE has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("RLP_LX_BYTESIZE has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("TYPE has not been filled");
      }

      filled.clear();

      return this;
    }

    TraceBuilder fillAndValidateRow() {
      if (!filled.get(20)) {
        absTxNum.add(BigInteger.ZERO);
        this.filled.set(20);
      }
      if (!filled.get(28)) {
        absTxNumInfiny.add(BigInteger.ZERO);
        this.filled.set(28);
      }
      if (!filled.get(56)) {
        acc1.add(BigInteger.ZERO);
        this.filled.set(56);
      }
      if (!filled.get(32)) {
        acc2.add(BigInteger.ZERO);
        this.filled.set(32);
      }
      if (!filled.get(10)) {
        accBytesize.add(BigInteger.ZERO);
        this.filled.set(10);
      }
      if (!filled.get(55)) {
        accessTupleBytesize.add(BigInteger.ZERO);
        this.filled.set(55);
      }
      if (!filled.get(46)) {
        addrHi.add(BigInteger.ZERO);
        this.filled.set(46);
      }
      if (!filled.get(22)) {
        addrLo.add(BigInteger.ZERO);
        this.filled.set(22);
      }
      if (!filled.get(13)) {
        bit.add(false);
        this.filled.set(13);
      }
      if (!filled.get(8)) {
        bitAcc.add(UnsignedByte.of(0));
        this.filled.set(8);
      }
      if (!filled.get(25)) {
        byte1.add(UnsignedByte.of(0));
        this.filled.set(25);
      }
      if (!filled.get(15)) {
        byte2.add(UnsignedByte.of(0));
        this.filled.set(15);
      }
      if (!filled.get(36)) {
        codeFragmentIndex.add(BigInteger.ZERO);
        this.filled.set(36);
      }
      if (!filled.get(35)) {
        counter.add(UnsignedByte.of(0));
        this.filled.set(35);
      }
      if (!filled.get(40)) {
        dataHi.add(BigInteger.ZERO);
        this.filled.set(40);
      }
      if (!filled.get(11)) {
        dataLo.add(BigInteger.ZERO);
        this.filled.set(11);
      }
      if (!filled.get(33)) {
        datagascost.add(BigInteger.ZERO);
        this.filled.set(33);
      }
      if (!filled.get(7)) {
        depth1.add(false);
        this.filled.set(7);
      }
      if (!filled.get(58)) {
        depth2.add(false);
        this.filled.set(58);
      }
      if (!filled.get(31)) {
        done.add(false);
        this.filled.set(31);
      }
      if (!filled.get(5)) {
        endPhase.add(false);
        this.filled.set(5);
      }
      if (!filled.get(39)) {
        indexData.add(BigInteger.ZERO);
        this.filled.set(39);
      }
      if (!filled.get(51)) {
        indexLt.add(BigInteger.ZERO);
        this.filled.set(51);
      }
      if (!filled.get(42)) {
        indexLx.add(BigInteger.ZERO);
        this.filled.set(42);
      }
      if (!filled.get(49)) {
        input1.add(BigInteger.ZERO);
        this.filled.set(49);
      }
      if (!filled.get(17)) {
        input2.add(BigInteger.ZERO);
        this.filled.set(17);
      }
      if (!filled.get(54)) {
        isPadding.add(false);
        this.filled.set(54);
      }
      if (!filled.get(38)) {
        isPrefix.add(false);
        this.filled.set(38);
      }
      if (!filled.get(4)) {
        limb.add(BigInteger.ZERO);
        this.filled.set(4);
      }
      if (!filled.get(27)) {
        limbConstructed.add(false);
        this.filled.set(27);
      }
      if (!filled.get(23)) {
        lt.add(false);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        lx.add(false);
        this.filled.set(24);
      }
      if (!filled.get(1)) {
        nBytes.add(UnsignedByte.of(0));
        this.filled.set(1);
      }
      if (!filled.get(50)) {
        nbAddr.add(BigInteger.ZERO);
        this.filled.set(50);
      }
      if (!filled.get(37)) {
        nbSto.add(BigInteger.ZERO);
        this.filled.set(37);
      }
      if (!filled.get(26)) {
        nbStoPerAddr.add(BigInteger.ZERO);
        this.filled.set(26);
      }
      if (!filled.get(6)) {
        numberStep.add(UnsignedByte.of(0));
        this.filled.set(6);
      }
      if (!filled.get(34)) {
        phase0.add(false);
        this.filled.set(34);
      }
      if (!filled.get(18)) {
        phase1.add(false);
        this.filled.set(18);
      }
      if (!filled.get(30)) {
        phase10.add(false);
        this.filled.set(30);
      }
      if (!filled.get(52)) {
        phase11.add(false);
        this.filled.set(52);
      }
      if (!filled.get(47)) {
        phase12.add(false);
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        phase13.add(false);
        this.filled.set(48);
      }
      if (!filled.get(60)) {
        phase14.add(false);
        this.filled.set(60);
      }
      if (!filled.get(9)) {
        phase2.add(false);
        this.filled.set(9);
      }
      if (!filled.get(2)) {
        phase3.add(false);
        this.filled.set(2);
      }
      if (!filled.get(12)) {
        phase4.add(false);
        this.filled.set(12);
      }
      if (!filled.get(3)) {
        phase5.add(false);
        this.filled.set(3);
      }
      if (!filled.get(41)) {
        phase6.add(false);
        this.filled.set(41);
      }
      if (!filled.get(57)) {
        phase7.add(false);
        this.filled.set(57);
      }
      if (!filled.get(19)) {
        phase8.add(false);
        this.filled.set(19);
      }
      if (!filled.get(16)) {
        phase9.add(false);
        this.filled.set(16);
      }
      if (!filled.get(21)) {
        phaseBytesize.add(BigInteger.ZERO);
        this.filled.set(21);
      }
      if (!filled.get(53)) {
        power.add(BigInteger.ZERO);
        this.filled.set(53);
      }
      if (!filled.get(0)) {
        requiresEvmExecution.add(false);
        this.filled.set(0);
      }
      if (!filled.get(59)) {
        rlpLtBytesize.add(BigInteger.ZERO);
        this.filled.set(59);
      }
      if (!filled.get(45)) {
        rlpLxBytesize.add(BigInteger.ZERO);
        this.filled.set(45);
      }
      if (!filled.get(43)) {
        type.add(UnsignedByte.of(0));
        this.filled.set(43);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          absTxNum,
          absTxNumInfiny,
          acc1,
          acc2,
          accBytesize,
          accessTupleBytesize,
          addrHi,
          addrLo,
          bit,
          bitAcc,
          byte1,
          byte2,
          codeFragmentIndex,
          counter,
          dataHi,
          dataLo,
          datagascost,
          depth1,
          depth2,
          done,
          endPhase,
          indexData,
          indexLt,
          indexLx,
          input1,
          input2,
          isPadding,
          isPrefix,
          limb,
          limbConstructed,
          lt,
          lx,
          nBytes,
          nbAddr,
          nbSto,
          nbStoPerAddr,
          numberStep,
          phase0,
          phase1,
          phase10,
          phase11,
          phase12,
          phase13,
          phase14,
          phase2,
          phase3,
          phase4,
          phase5,
          phase6,
          phase7,
          phase8,
          phase9,
          phaseBytesize,
          power,
          requiresEvmExecution,
          rlpLtBytesize,
          rlpLxBytesize,
          type);
    }
  }
}
