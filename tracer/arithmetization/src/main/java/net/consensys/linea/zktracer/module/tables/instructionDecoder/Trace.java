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

package net.consensys.linea.zktracer.module.tables.instructionDecoder;

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
    @JsonProperty("ADDRESS_TRIMMING_INSTRUCTION") List<Boolean> addressTrimmingInstruction,
    @JsonProperty("ALPHA") List<UnsignedByte> alpha,
    @JsonProperty("BILLING_PER_BYTE") List<BigInteger> billingPerByte,
    @JsonProperty("BILLING_PER_WORD") List<BigInteger> billingPerWord,
    @JsonProperty("DELTA") List<UnsignedByte> delta,
    @JsonProperty("FAMILY_ACCOUNT") List<Boolean> familyAccount,
    @JsonProperty("FAMILY_ADD") List<Boolean> familyAdd,
    @JsonProperty("FAMILY_BATCH") List<Boolean> familyBatch,
    @JsonProperty("FAMILY_BIN") List<Boolean> familyBin,
    @JsonProperty("FAMILY_CALL") List<Boolean> familyCall,
    @JsonProperty("FAMILY_CONTEXT") List<Boolean> familyContext,
    @JsonProperty("FAMILY_COPY") List<Boolean> familyCopy,
    @JsonProperty("FAMILY_CREATE") List<Boolean> familyCreate,
    @JsonProperty("FAMILY_DUP") List<Boolean> familyDup,
    @JsonProperty("FAMILY_EXT") List<Boolean> familyExt,
    @JsonProperty("FAMILY_HALT") List<Boolean> familyHalt,
    @JsonProperty("FAMILY_INVALID") List<Boolean> familyInvalid,
    @JsonProperty("FAMILY_JUMP") List<Boolean> familyJump,
    @JsonProperty("FAMILY_KEC") List<Boolean> familyKec,
    @JsonProperty("FAMILY_LOG") List<Boolean> familyLog,
    @JsonProperty("FAMILY_MACHINE_STATE") List<Boolean> familyMachineState,
    @JsonProperty("FAMILY_MOD") List<Boolean> familyMod,
    @JsonProperty("FAMILY_MUL") List<Boolean> familyMul,
    @JsonProperty("FAMILY_PUSH_POP") List<Boolean> familyPushPop,
    @JsonProperty("FAMILY_SHF") List<Boolean> familyShf,
    @JsonProperty("FAMILY_STACK_RAM") List<Boolean> familyStackRam,
    @JsonProperty("FAMILY_STORAGE") List<Boolean> familyStorage,
    @JsonProperty("FAMILY_SWAP") List<Boolean> familySwap,
    @JsonProperty("FAMILY_TRANSACTION") List<Boolean> familyTransaction,
    @JsonProperty("FAMILY_WCP") List<Boolean> familyWcp,
    @JsonProperty("FLAG1") List<Boolean> flag1,
    @JsonProperty("FLAG2") List<Boolean> flag2,
    @JsonProperty("FLAG3") List<Boolean> flag3,
    @JsonProperty("FLAG4") List<Boolean> flag4,
    @JsonProperty("FORBIDDEN_IN_STATIC") List<Boolean> forbiddenInStatic,
    @JsonProperty("IS_JUMPDEST") List<Boolean> isJumpdest,
    @JsonProperty("IS_PUSH") List<Boolean> isPush,
    @JsonProperty("MXP_TYPE_1") List<Boolean> mxpType1,
    @JsonProperty("MXP_TYPE_2") List<Boolean> mxpType2,
    @JsonProperty("MXP_TYPE_3") List<Boolean> mxpType3,
    @JsonProperty("MXP_TYPE_4") List<Boolean> mxpType4,
    @JsonProperty("MXP_TYPE_5") List<Boolean> mxpType5,
    @JsonProperty("NB_ADDED") List<UnsignedByte> nbAdded,
    @JsonProperty("NB_REMOVED") List<UnsignedByte> nbRemoved,
    @JsonProperty("OPCODE") List<BigInteger> opcode,
    @JsonProperty("PATTERN_CALL") List<Boolean> patternCall,
    @JsonProperty("PATTERN_COPY") List<Boolean> patternCopy,
    @JsonProperty("PATTERN_CREATE") List<Boolean> patternCreate,
    @JsonProperty("PATTERN_DUP") List<Boolean> patternDup,
    @JsonProperty("PATTERN_LOAD_STORE") List<Boolean> patternLoadStore,
    @JsonProperty("PATTERN_LOG") List<Boolean> patternLog,
    @JsonProperty("PATTERN_ONE_ONE") List<Boolean> patternOneOne,
    @JsonProperty("PATTERN_ONE_ZERO") List<Boolean> patternOneZero,
    @JsonProperty("PATTERN_SWAP") List<Boolean> patternSwap,
    @JsonProperty("PATTERN_THREE_ONE") List<Boolean> patternThreeOne,
    @JsonProperty("PATTERN_TWO_ONE") List<Boolean> patternTwoOne,
    @JsonProperty("PATTERN_TWO_ZERO") List<Boolean> patternTwoZero,
    @JsonProperty("PATTERN_ZERO_ONE") List<Boolean> patternZeroOne,
    @JsonProperty("PATTERN_ZERO_ZERO") List<Boolean> patternZeroZero,
    @JsonProperty("RAM_ENABLED") List<Boolean> ramEnabled,
    @JsonProperty("RAM_SOURCE_BLAKE_DATA") List<Boolean> ramSourceBlakeData,
    @JsonProperty("RAM_SOURCE_EC_DATA") List<Boolean> ramSourceEcData,
    @JsonProperty("RAM_SOURCE_EC_INFO") List<Boolean> ramSourceEcInfo,
    @JsonProperty("RAM_SOURCE_HASH_DATA") List<Boolean> ramSourceHashData,
    @JsonProperty("RAM_SOURCE_HASH_INFO") List<Boolean> ramSourceHashInfo,
    @JsonProperty("RAM_SOURCE_LOG_DATA") List<Boolean> ramSourceLogData,
    @JsonProperty("RAM_SOURCE_MODEXP_DATA") List<Boolean> ramSourceModexpData,
    @JsonProperty("RAM_SOURCE_RAM") List<Boolean> ramSourceRam,
    @JsonProperty("RAM_SOURCE_ROM") List<Boolean> ramSourceRom,
    @JsonProperty("RAM_SOURCE_STACK") List<Boolean> ramSourceStack,
    @JsonProperty("RAM_SOURCE_TXN_DATA") List<Boolean> ramSourceTxnData,
    @JsonProperty("RAM_TARGET_BLAKE_DATA") List<Boolean> ramTargetBlakeData,
    @JsonProperty("RAM_TARGET_EC_DATA") List<Boolean> ramTargetEcData,
    @JsonProperty("RAM_TARGET_EC_INFO") List<Boolean> ramTargetEcInfo,
    @JsonProperty("RAM_TARGET_HASH_DATA") List<Boolean> ramTargetHashData,
    @JsonProperty("RAM_TARGET_HASH_INFO") List<Boolean> ramTargetHashInfo,
    @JsonProperty("RAM_TARGET_LOG_DATA") List<Boolean> ramTargetLogData,
    @JsonProperty("RAM_TARGET_MODEXP_DATA") List<Boolean> ramTargetModexpData,
    @JsonProperty("RAM_TARGET_RAM") List<Boolean> ramTargetRam,
    @JsonProperty("RAM_TARGET_ROM") List<Boolean> ramTargetRom,
    @JsonProperty("RAM_TARGET_STACK") List<Boolean> ramTargetStack,
    @JsonProperty("RAM_TARGET_TXN_DATA") List<Boolean> ramTargetTxnData,
    @JsonProperty("STATIC_GAS") List<BigInteger> staticGas,
    @JsonProperty("TWO_LINES_INSTRUCTION") List<Boolean> twoLinesInstruction) {
  static TraceBuilder builder(int length) {
    return new TraceBuilder(length);
  }

  public int size() {
    return this.addressTrimmingInstruction.size();
  }

  static class TraceBuilder {
    private final BitSet filled = new BitSet();

    @JsonProperty("ADDRESS_TRIMMING_INSTRUCTION")
    private final List<Boolean> addressTrimmingInstruction;

    @JsonProperty("ALPHA")
    private final List<UnsignedByte> alpha;

    @JsonProperty("BILLING_PER_BYTE")
    private final List<BigInteger> billingPerByte;

    @JsonProperty("BILLING_PER_WORD")
    private final List<BigInteger> billingPerWord;

    @JsonProperty("DELTA")
    private final List<UnsignedByte> delta;

    @JsonProperty("FAMILY_ACCOUNT")
    private final List<Boolean> familyAccount;

    @JsonProperty("FAMILY_ADD")
    private final List<Boolean> familyAdd;

    @JsonProperty("FAMILY_BATCH")
    private final List<Boolean> familyBatch;

    @JsonProperty("FAMILY_BIN")
    private final List<Boolean> familyBin;

    @JsonProperty("FAMILY_CALL")
    private final List<Boolean> familyCall;

    @JsonProperty("FAMILY_CONTEXT")
    private final List<Boolean> familyContext;

    @JsonProperty("FAMILY_COPY")
    private final List<Boolean> familyCopy;

    @JsonProperty("FAMILY_CREATE")
    private final List<Boolean> familyCreate;

    @JsonProperty("FAMILY_DUP")
    private final List<Boolean> familyDup;

    @JsonProperty("FAMILY_EXT")
    private final List<Boolean> familyExt;

    @JsonProperty("FAMILY_HALT")
    private final List<Boolean> familyHalt;

    @JsonProperty("FAMILY_INVALID")
    private final List<Boolean> familyInvalid;

    @JsonProperty("FAMILY_JUMP")
    private final List<Boolean> familyJump;

    @JsonProperty("FAMILY_KEC")
    private final List<Boolean> familyKec;

    @JsonProperty("FAMILY_LOG")
    private final List<Boolean> familyLog;

    @JsonProperty("FAMILY_MACHINE_STATE")
    private final List<Boolean> familyMachineState;

    @JsonProperty("FAMILY_MOD")
    private final List<Boolean> familyMod;

    @JsonProperty("FAMILY_MUL")
    private final List<Boolean> familyMul;

    @JsonProperty("FAMILY_PUSH_POP")
    private final List<Boolean> familyPushPop;

    @JsonProperty("FAMILY_SHF")
    private final List<Boolean> familyShf;

    @JsonProperty("FAMILY_STACK_RAM")
    private final List<Boolean> familyStackRam;

    @JsonProperty("FAMILY_STORAGE")
    private final List<Boolean> familyStorage;

    @JsonProperty("FAMILY_SWAP")
    private final List<Boolean> familySwap;

    @JsonProperty("FAMILY_TRANSACTION")
    private final List<Boolean> familyTransaction;

    @JsonProperty("FAMILY_WCP")
    private final List<Boolean> familyWcp;

    @JsonProperty("FLAG1")
    private final List<Boolean> flag1;

    @JsonProperty("FLAG2")
    private final List<Boolean> flag2;

    @JsonProperty("FLAG3")
    private final List<Boolean> flag3;

    @JsonProperty("FLAG4")
    private final List<Boolean> flag4;

    @JsonProperty("FORBIDDEN_IN_STATIC")
    private final List<Boolean> forbiddenInStatic;

    @JsonProperty("IS_JUMPDEST")
    private final List<Boolean> isJumpdest;

    @JsonProperty("IS_PUSH")
    private final List<Boolean> isPush;

    @JsonProperty("MXP_TYPE_1")
    private final List<Boolean> mxpType1;

    @JsonProperty("MXP_TYPE_2")
    private final List<Boolean> mxpType2;

    @JsonProperty("MXP_TYPE_3")
    private final List<Boolean> mxpType3;

    @JsonProperty("MXP_TYPE_4")
    private final List<Boolean> mxpType4;

    @JsonProperty("MXP_TYPE_5")
    private final List<Boolean> mxpType5;

    @JsonProperty("NB_ADDED")
    private final List<UnsignedByte> nbAdded;

    @JsonProperty("NB_REMOVED")
    private final List<UnsignedByte> nbRemoved;

    @JsonProperty("OPCODE")
    private final List<BigInteger> opcode;

    @JsonProperty("PATTERN_CALL")
    private final List<Boolean> patternCall;

    @JsonProperty("PATTERN_COPY")
    private final List<Boolean> patternCopy;

    @JsonProperty("PATTERN_CREATE")
    private final List<Boolean> patternCreate;

    @JsonProperty("PATTERN_DUP")
    private final List<Boolean> patternDup;

    @JsonProperty("PATTERN_LOAD_STORE")
    private final List<Boolean> patternLoadStore;

    @JsonProperty("PATTERN_LOG")
    private final List<Boolean> patternLog;

    @JsonProperty("PATTERN_ONE_ONE")
    private final List<Boolean> patternOneOne;

    @JsonProperty("PATTERN_ONE_ZERO")
    private final List<Boolean> patternOneZero;

    @JsonProperty("PATTERN_SWAP")
    private final List<Boolean> patternSwap;

    @JsonProperty("PATTERN_THREE_ONE")
    private final List<Boolean> patternThreeOne;

    @JsonProperty("PATTERN_TWO_ONE")
    private final List<Boolean> patternTwoOne;

    @JsonProperty("PATTERN_TWO_ZERO")
    private final List<Boolean> patternTwoZero;

    @JsonProperty("PATTERN_ZERO_ONE")
    private final List<Boolean> patternZeroOne;

    @JsonProperty("PATTERN_ZERO_ZERO")
    private final List<Boolean> patternZeroZero;

    @JsonProperty("RAM_ENABLED")
    private final List<Boolean> ramEnabled;

    @JsonProperty("RAM_SOURCE_BLAKE_DATA")
    private final List<Boolean> ramSourceBlakeData;

    @JsonProperty("RAM_SOURCE_EC_DATA")
    private final List<Boolean> ramSourceEcData;

    @JsonProperty("RAM_SOURCE_EC_INFO")
    private final List<Boolean> ramSourceEcInfo;

    @JsonProperty("RAM_SOURCE_HASH_DATA")
    private final List<Boolean> ramSourceHashData;

    @JsonProperty("RAM_SOURCE_HASH_INFO")
    private final List<Boolean> ramSourceHashInfo;

    @JsonProperty("RAM_SOURCE_LOG_DATA")
    private final List<Boolean> ramSourceLogData;

    @JsonProperty("RAM_SOURCE_MODEXP_DATA")
    private final List<Boolean> ramSourceModexpData;

    @JsonProperty("RAM_SOURCE_RAM")
    private final List<Boolean> ramSourceRam;

    @JsonProperty("RAM_SOURCE_ROM")
    private final List<Boolean> ramSourceRom;

    @JsonProperty("RAM_SOURCE_STACK")
    private final List<Boolean> ramSourceStack;

    @JsonProperty("RAM_SOURCE_TXN_DATA")
    private final List<Boolean> ramSourceTxnData;

    @JsonProperty("RAM_TARGET_BLAKE_DATA")
    private final List<Boolean> ramTargetBlakeData;

    @JsonProperty("RAM_TARGET_EC_DATA")
    private final List<Boolean> ramTargetEcData;

    @JsonProperty("RAM_TARGET_EC_INFO")
    private final List<Boolean> ramTargetEcInfo;

    @JsonProperty("RAM_TARGET_HASH_DATA")
    private final List<Boolean> ramTargetHashData;

    @JsonProperty("RAM_TARGET_HASH_INFO")
    private final List<Boolean> ramTargetHashInfo;

    @JsonProperty("RAM_TARGET_LOG_DATA")
    private final List<Boolean> ramTargetLogData;

    @JsonProperty("RAM_TARGET_MODEXP_DATA")
    private final List<Boolean> ramTargetModexpData;

    @JsonProperty("RAM_TARGET_RAM")
    private final List<Boolean> ramTargetRam;

    @JsonProperty("RAM_TARGET_ROM")
    private final List<Boolean> ramTargetRom;

    @JsonProperty("RAM_TARGET_STACK")
    private final List<Boolean> ramTargetStack;

    @JsonProperty("RAM_TARGET_TXN_DATA")
    private final List<Boolean> ramTargetTxnData;

    @JsonProperty("STATIC_GAS")
    private final List<BigInteger> staticGas;

    @JsonProperty("TWO_LINES_INSTRUCTION")
    private final List<Boolean> twoLinesInstruction;

    TraceBuilder(int length) {
      this.addressTrimmingInstruction = new ArrayList<>(length);
      this.alpha = new ArrayList<>(length);
      this.billingPerByte = new ArrayList<>(length);
      this.billingPerWord = new ArrayList<>(length);
      this.delta = new ArrayList<>(length);
      this.familyAccount = new ArrayList<>(length);
      this.familyAdd = new ArrayList<>(length);
      this.familyBatch = new ArrayList<>(length);
      this.familyBin = new ArrayList<>(length);
      this.familyCall = new ArrayList<>(length);
      this.familyContext = new ArrayList<>(length);
      this.familyCopy = new ArrayList<>(length);
      this.familyCreate = new ArrayList<>(length);
      this.familyDup = new ArrayList<>(length);
      this.familyExt = new ArrayList<>(length);
      this.familyHalt = new ArrayList<>(length);
      this.familyInvalid = new ArrayList<>(length);
      this.familyJump = new ArrayList<>(length);
      this.familyKec = new ArrayList<>(length);
      this.familyLog = new ArrayList<>(length);
      this.familyMachineState = new ArrayList<>(length);
      this.familyMod = new ArrayList<>(length);
      this.familyMul = new ArrayList<>(length);
      this.familyPushPop = new ArrayList<>(length);
      this.familyShf = new ArrayList<>(length);
      this.familyStackRam = new ArrayList<>(length);
      this.familyStorage = new ArrayList<>(length);
      this.familySwap = new ArrayList<>(length);
      this.familyTransaction = new ArrayList<>(length);
      this.familyWcp = new ArrayList<>(length);
      this.flag1 = new ArrayList<>(length);
      this.flag2 = new ArrayList<>(length);
      this.flag3 = new ArrayList<>(length);
      this.flag4 = new ArrayList<>(length);
      this.forbiddenInStatic = new ArrayList<>(length);
      this.isJumpdest = new ArrayList<>(length);
      this.isPush = new ArrayList<>(length);
      this.mxpType1 = new ArrayList<>(length);
      this.mxpType2 = new ArrayList<>(length);
      this.mxpType3 = new ArrayList<>(length);
      this.mxpType4 = new ArrayList<>(length);
      this.mxpType5 = new ArrayList<>(length);
      this.nbAdded = new ArrayList<>(length);
      this.nbRemoved = new ArrayList<>(length);
      this.opcode = new ArrayList<>(length);
      this.patternCall = new ArrayList<>(length);
      this.patternCopy = new ArrayList<>(length);
      this.patternCreate = new ArrayList<>(length);
      this.patternDup = new ArrayList<>(length);
      this.patternLoadStore = new ArrayList<>(length);
      this.patternLog = new ArrayList<>(length);
      this.patternOneOne = new ArrayList<>(length);
      this.patternOneZero = new ArrayList<>(length);
      this.patternSwap = new ArrayList<>(length);
      this.patternThreeOne = new ArrayList<>(length);
      this.patternTwoOne = new ArrayList<>(length);
      this.patternTwoZero = new ArrayList<>(length);
      this.patternZeroOne = new ArrayList<>(length);
      this.patternZeroZero = new ArrayList<>(length);
      this.ramEnabled = new ArrayList<>(length);
      this.ramSourceBlakeData = new ArrayList<>(length);
      this.ramSourceEcData = new ArrayList<>(length);
      this.ramSourceEcInfo = new ArrayList<>(length);
      this.ramSourceHashData = new ArrayList<>(length);
      this.ramSourceHashInfo = new ArrayList<>(length);
      this.ramSourceLogData = new ArrayList<>(length);
      this.ramSourceModexpData = new ArrayList<>(length);
      this.ramSourceRam = new ArrayList<>(length);
      this.ramSourceRom = new ArrayList<>(length);
      this.ramSourceStack = new ArrayList<>(length);
      this.ramSourceTxnData = new ArrayList<>(length);
      this.ramTargetBlakeData = new ArrayList<>(length);
      this.ramTargetEcData = new ArrayList<>(length);
      this.ramTargetEcInfo = new ArrayList<>(length);
      this.ramTargetHashData = new ArrayList<>(length);
      this.ramTargetHashInfo = new ArrayList<>(length);
      this.ramTargetLogData = new ArrayList<>(length);
      this.ramTargetModexpData = new ArrayList<>(length);
      this.ramTargetRam = new ArrayList<>(length);
      this.ramTargetRom = new ArrayList<>(length);
      this.ramTargetStack = new ArrayList<>(length);
      this.ramTargetTxnData = new ArrayList<>(length);
      this.staticGas = new ArrayList<>(length);
      this.twoLinesInstruction = new ArrayList<>(length);
    }

    public int size() {
      if (!filled.isEmpty()) {
        throw new RuntimeException("Cannot measure a trace with a non-validated row.");
      }

      return this.addressTrimmingInstruction.size();
    }

    public TraceBuilder addressTrimmingInstruction(final Boolean b) {
      if (filled.get(0)) {
        throw new IllegalStateException("ADDRESS_TRIMMING_INSTRUCTION already set");
      } else {
        filled.set(0);
      }

      addressTrimmingInstruction.add(b);

      return this;
    }

    public TraceBuilder alpha(final UnsignedByte b) {
      if (filled.get(1)) {
        throw new IllegalStateException("ALPHA already set");
      } else {
        filled.set(1);
      }

      alpha.add(b);

      return this;
    }

    public TraceBuilder billingPerByte(final BigInteger b) {
      if (filled.get(2)) {
        throw new IllegalStateException("BILLING_PER_BYTE already set");
      } else {
        filled.set(2);
      }

      billingPerByte.add(b);

      return this;
    }

    public TraceBuilder billingPerWord(final BigInteger b) {
      if (filled.get(3)) {
        throw new IllegalStateException("BILLING_PER_WORD already set");
      } else {
        filled.set(3);
      }

      billingPerWord.add(b);

      return this;
    }

    public TraceBuilder delta(final UnsignedByte b) {
      if (filled.get(4)) {
        throw new IllegalStateException("DELTA already set");
      } else {
        filled.set(4);
      }

      delta.add(b);

      return this;
    }

    public TraceBuilder familyAccount(final Boolean b) {
      if (filled.get(5)) {
        throw new IllegalStateException("FAMILY_ACCOUNT already set");
      } else {
        filled.set(5);
      }

      familyAccount.add(b);

      return this;
    }

    public TraceBuilder familyAdd(final Boolean b) {
      if (filled.get(6)) {
        throw new IllegalStateException("FAMILY_ADD already set");
      } else {
        filled.set(6);
      }

      familyAdd.add(b);

      return this;
    }

    public TraceBuilder familyBatch(final Boolean b) {
      if (filled.get(7)) {
        throw new IllegalStateException("FAMILY_BATCH already set");
      } else {
        filled.set(7);
      }

      familyBatch.add(b);

      return this;
    }

    public TraceBuilder familyBin(final Boolean b) {
      if (filled.get(8)) {
        throw new IllegalStateException("FAMILY_BIN already set");
      } else {
        filled.set(8);
      }

      familyBin.add(b);

      return this;
    }

    public TraceBuilder familyCall(final Boolean b) {
      if (filled.get(9)) {
        throw new IllegalStateException("FAMILY_CALL already set");
      } else {
        filled.set(9);
      }

      familyCall.add(b);

      return this;
    }

    public TraceBuilder familyContext(final Boolean b) {
      if (filled.get(10)) {
        throw new IllegalStateException("FAMILY_CONTEXT already set");
      } else {
        filled.set(10);
      }

      familyContext.add(b);

      return this;
    }

    public TraceBuilder familyCopy(final Boolean b) {
      if (filled.get(11)) {
        throw new IllegalStateException("FAMILY_COPY already set");
      } else {
        filled.set(11);
      }

      familyCopy.add(b);

      return this;
    }

    public TraceBuilder familyCreate(final Boolean b) {
      if (filled.get(12)) {
        throw new IllegalStateException("FAMILY_CREATE already set");
      } else {
        filled.set(12);
      }

      familyCreate.add(b);

      return this;
    }

    public TraceBuilder familyDup(final Boolean b) {
      if (filled.get(13)) {
        throw new IllegalStateException("FAMILY_DUP already set");
      } else {
        filled.set(13);
      }

      familyDup.add(b);

      return this;
    }

    public TraceBuilder familyExt(final Boolean b) {
      if (filled.get(14)) {
        throw new IllegalStateException("FAMILY_EXT already set");
      } else {
        filled.set(14);
      }

      familyExt.add(b);

      return this;
    }

    public TraceBuilder familyHalt(final Boolean b) {
      if (filled.get(15)) {
        throw new IllegalStateException("FAMILY_HALT already set");
      } else {
        filled.set(15);
      }

      familyHalt.add(b);

      return this;
    }

    public TraceBuilder familyInvalid(final Boolean b) {
      if (filled.get(16)) {
        throw new IllegalStateException("FAMILY_INVALID already set");
      } else {
        filled.set(16);
      }

      familyInvalid.add(b);

      return this;
    }

    public TraceBuilder familyJump(final Boolean b) {
      if (filled.get(17)) {
        throw new IllegalStateException("FAMILY_JUMP already set");
      } else {
        filled.set(17);
      }

      familyJump.add(b);

      return this;
    }

    public TraceBuilder familyKec(final Boolean b) {
      if (filled.get(18)) {
        throw new IllegalStateException("FAMILY_KEC already set");
      } else {
        filled.set(18);
      }

      familyKec.add(b);

      return this;
    }

    public TraceBuilder familyLog(final Boolean b) {
      if (filled.get(19)) {
        throw new IllegalStateException("FAMILY_LOG already set");
      } else {
        filled.set(19);
      }

      familyLog.add(b);

      return this;
    }

    public TraceBuilder familyMachineState(final Boolean b) {
      if (filled.get(20)) {
        throw new IllegalStateException("FAMILY_MACHINE_STATE already set");
      } else {
        filled.set(20);
      }

      familyMachineState.add(b);

      return this;
    }

    public TraceBuilder familyMod(final Boolean b) {
      if (filled.get(21)) {
        throw new IllegalStateException("FAMILY_MOD already set");
      } else {
        filled.set(21);
      }

      familyMod.add(b);

      return this;
    }

    public TraceBuilder familyMul(final Boolean b) {
      if (filled.get(22)) {
        throw new IllegalStateException("FAMILY_MUL already set");
      } else {
        filled.set(22);
      }

      familyMul.add(b);

      return this;
    }

    public TraceBuilder familyPushPop(final Boolean b) {
      if (filled.get(23)) {
        throw new IllegalStateException("FAMILY_PUSH_POP already set");
      } else {
        filled.set(23);
      }

      familyPushPop.add(b);

      return this;
    }

    public TraceBuilder familyShf(final Boolean b) {
      if (filled.get(24)) {
        throw new IllegalStateException("FAMILY_SHF already set");
      } else {
        filled.set(24);
      }

      familyShf.add(b);

      return this;
    }

    public TraceBuilder familyStackRam(final Boolean b) {
      if (filled.get(25)) {
        throw new IllegalStateException("FAMILY_STACK_RAM already set");
      } else {
        filled.set(25);
      }

      familyStackRam.add(b);

      return this;
    }

    public TraceBuilder familyStorage(final Boolean b) {
      if (filled.get(26)) {
        throw new IllegalStateException("FAMILY_STORAGE already set");
      } else {
        filled.set(26);
      }

      familyStorage.add(b);

      return this;
    }

    public TraceBuilder familySwap(final Boolean b) {
      if (filled.get(27)) {
        throw new IllegalStateException("FAMILY_SWAP already set");
      } else {
        filled.set(27);
      }

      familySwap.add(b);

      return this;
    }

    public TraceBuilder familyTransaction(final Boolean b) {
      if (filled.get(28)) {
        throw new IllegalStateException("FAMILY_TRANSACTION already set");
      } else {
        filled.set(28);
      }

      familyTransaction.add(b);

      return this;
    }

    public TraceBuilder familyWcp(final Boolean b) {
      if (filled.get(29)) {
        throw new IllegalStateException("FAMILY_WCP already set");
      } else {
        filled.set(29);
      }

      familyWcp.add(b);

      return this;
    }

    public TraceBuilder flag1(final Boolean b) {
      if (filled.get(30)) {
        throw new IllegalStateException("FLAG1 already set");
      } else {
        filled.set(30);
      }

      flag1.add(b);

      return this;
    }

    public TraceBuilder flag2(final Boolean b) {
      if (filled.get(31)) {
        throw new IllegalStateException("FLAG2 already set");
      } else {
        filled.set(31);
      }

      flag2.add(b);

      return this;
    }

    public TraceBuilder flag3(final Boolean b) {
      if (filled.get(32)) {
        throw new IllegalStateException("FLAG3 already set");
      } else {
        filled.set(32);
      }

      flag3.add(b);

      return this;
    }

    public TraceBuilder flag4(final Boolean b) {
      if (filled.get(33)) {
        throw new IllegalStateException("FLAG4 already set");
      } else {
        filled.set(33);
      }

      flag4.add(b);

      return this;
    }

    public TraceBuilder forbiddenInStatic(final Boolean b) {
      if (filled.get(34)) {
        throw new IllegalStateException("FORBIDDEN_IN_STATIC already set");
      } else {
        filled.set(34);
      }

      forbiddenInStatic.add(b);

      return this;
    }

    public TraceBuilder isJumpdest(final Boolean b) {
      if (filled.get(35)) {
        throw new IllegalStateException("IS_JUMPDEST already set");
      } else {
        filled.set(35);
      }

      isJumpdest.add(b);

      return this;
    }

    public TraceBuilder isPush(final Boolean b) {
      if (filled.get(36)) {
        throw new IllegalStateException("IS_PUSH already set");
      } else {
        filled.set(36);
      }

      isPush.add(b);

      return this;
    }

    public TraceBuilder mxpType1(final Boolean b) {
      if (filled.get(37)) {
        throw new IllegalStateException("MXP_TYPE_1 already set");
      } else {
        filled.set(37);
      }

      mxpType1.add(b);

      return this;
    }

    public TraceBuilder mxpType2(final Boolean b) {
      if (filled.get(38)) {
        throw new IllegalStateException("MXP_TYPE_2 already set");
      } else {
        filled.set(38);
      }

      mxpType2.add(b);

      return this;
    }

    public TraceBuilder mxpType3(final Boolean b) {
      if (filled.get(39)) {
        throw new IllegalStateException("MXP_TYPE_3 already set");
      } else {
        filled.set(39);
      }

      mxpType3.add(b);

      return this;
    }

    public TraceBuilder mxpType4(final Boolean b) {
      if (filled.get(40)) {
        throw new IllegalStateException("MXP_TYPE_4 already set");
      } else {
        filled.set(40);
      }

      mxpType4.add(b);

      return this;
    }

    public TraceBuilder mxpType5(final Boolean b) {
      if (filled.get(41)) {
        throw new IllegalStateException("MXP_TYPE_5 already set");
      } else {
        filled.set(41);
      }

      mxpType5.add(b);

      return this;
    }

    public TraceBuilder nbAdded(final UnsignedByte b) {
      if (filled.get(42)) {
        throw new IllegalStateException("NB_ADDED already set");
      } else {
        filled.set(42);
      }

      nbAdded.add(b);

      return this;
    }

    public TraceBuilder nbRemoved(final UnsignedByte b) {
      if (filled.get(43)) {
        throw new IllegalStateException("NB_REMOVED already set");
      } else {
        filled.set(43);
      }

      nbRemoved.add(b);

      return this;
    }

    public TraceBuilder opcode(final BigInteger b) {
      if (filled.get(44)) {
        throw new IllegalStateException("OPCODE already set");
      } else {
        filled.set(44);
      }

      opcode.add(b);

      return this;
    }

    public TraceBuilder patternCall(final Boolean b) {
      if (filled.get(45)) {
        throw new IllegalStateException("PATTERN_CALL already set");
      } else {
        filled.set(45);
      }

      patternCall.add(b);

      return this;
    }

    public TraceBuilder patternCopy(final Boolean b) {
      if (filled.get(46)) {
        throw new IllegalStateException("PATTERN_COPY already set");
      } else {
        filled.set(46);
      }

      patternCopy.add(b);

      return this;
    }

    public TraceBuilder patternCreate(final Boolean b) {
      if (filled.get(47)) {
        throw new IllegalStateException("PATTERN_CREATE already set");
      } else {
        filled.set(47);
      }

      patternCreate.add(b);

      return this;
    }

    public TraceBuilder patternDup(final Boolean b) {
      if (filled.get(48)) {
        throw new IllegalStateException("PATTERN_DUP already set");
      } else {
        filled.set(48);
      }

      patternDup.add(b);

      return this;
    }

    public TraceBuilder patternLoadStore(final Boolean b) {
      if (filled.get(49)) {
        throw new IllegalStateException("PATTERN_LOAD_STORE already set");
      } else {
        filled.set(49);
      }

      patternLoadStore.add(b);

      return this;
    }

    public TraceBuilder patternLog(final Boolean b) {
      if (filled.get(50)) {
        throw new IllegalStateException("PATTERN_LOG already set");
      } else {
        filled.set(50);
      }

      patternLog.add(b);

      return this;
    }

    public TraceBuilder patternOneOne(final Boolean b) {
      if (filled.get(51)) {
        throw new IllegalStateException("PATTERN_ONE_ONE already set");
      } else {
        filled.set(51);
      }

      patternOneOne.add(b);

      return this;
    }

    public TraceBuilder patternOneZero(final Boolean b) {
      if (filled.get(52)) {
        throw new IllegalStateException("PATTERN_ONE_ZERO already set");
      } else {
        filled.set(52);
      }

      patternOneZero.add(b);

      return this;
    }

    public TraceBuilder patternSwap(final Boolean b) {
      if (filled.get(53)) {
        throw new IllegalStateException("PATTERN_SWAP already set");
      } else {
        filled.set(53);
      }

      patternSwap.add(b);

      return this;
    }

    public TraceBuilder patternThreeOne(final Boolean b) {
      if (filled.get(54)) {
        throw new IllegalStateException("PATTERN_THREE_ONE already set");
      } else {
        filled.set(54);
      }

      patternThreeOne.add(b);

      return this;
    }

    public TraceBuilder patternTwoOne(final Boolean b) {
      if (filled.get(55)) {
        throw new IllegalStateException("PATTERN_TWO_ONE already set");
      } else {
        filled.set(55);
      }

      patternTwoOne.add(b);

      return this;
    }

    public TraceBuilder patternTwoZero(final Boolean b) {
      if (filled.get(56)) {
        throw new IllegalStateException("PATTERN_TWO_ZERO already set");
      } else {
        filled.set(56);
      }

      patternTwoZero.add(b);

      return this;
    }

    public TraceBuilder patternZeroOne(final Boolean b) {
      if (filled.get(57)) {
        throw new IllegalStateException("PATTERN_ZERO_ONE already set");
      } else {
        filled.set(57);
      }

      patternZeroOne.add(b);

      return this;
    }

    public TraceBuilder patternZeroZero(final Boolean b) {
      if (filled.get(58)) {
        throw new IllegalStateException("PATTERN_ZERO_ZERO already set");
      } else {
        filled.set(58);
      }

      patternZeroZero.add(b);

      return this;
    }

    public TraceBuilder ramEnabled(final Boolean b) {
      if (filled.get(59)) {
        throw new IllegalStateException("RAM_ENABLED already set");
      } else {
        filled.set(59);
      }

      ramEnabled.add(b);

      return this;
    }

    public TraceBuilder ramSourceBlakeData(final Boolean b) {
      if (filled.get(60)) {
        throw new IllegalStateException("RAM_SOURCE_BLAKE_DATA already set");
      } else {
        filled.set(60);
      }

      ramSourceBlakeData.add(b);

      return this;
    }

    public TraceBuilder ramSourceEcData(final Boolean b) {
      if (filled.get(61)) {
        throw new IllegalStateException("RAM_SOURCE_EC_DATA already set");
      } else {
        filled.set(61);
      }

      ramSourceEcData.add(b);

      return this;
    }

    public TraceBuilder ramSourceEcInfo(final Boolean b) {
      if (filled.get(62)) {
        throw new IllegalStateException("RAM_SOURCE_EC_INFO already set");
      } else {
        filled.set(62);
      }

      ramSourceEcInfo.add(b);

      return this;
    }

    public TraceBuilder ramSourceHashData(final Boolean b) {
      if (filled.get(63)) {
        throw new IllegalStateException("RAM_SOURCE_HASH_DATA already set");
      } else {
        filled.set(63);
      }

      ramSourceHashData.add(b);

      return this;
    }

    public TraceBuilder ramSourceHashInfo(final Boolean b) {
      if (filled.get(64)) {
        throw new IllegalStateException("RAM_SOURCE_HASH_INFO already set");
      } else {
        filled.set(64);
      }

      ramSourceHashInfo.add(b);

      return this;
    }

    public TraceBuilder ramSourceLogData(final Boolean b) {
      if (filled.get(65)) {
        throw new IllegalStateException("RAM_SOURCE_LOG_DATA already set");
      } else {
        filled.set(65);
      }

      ramSourceLogData.add(b);

      return this;
    }

    public TraceBuilder ramSourceModexpData(final Boolean b) {
      if (filled.get(66)) {
        throw new IllegalStateException("RAM_SOURCE_MODEXP_DATA already set");
      } else {
        filled.set(66);
      }

      ramSourceModexpData.add(b);

      return this;
    }

    public TraceBuilder ramSourceRam(final Boolean b) {
      if (filled.get(67)) {
        throw new IllegalStateException("RAM_SOURCE_RAM already set");
      } else {
        filled.set(67);
      }

      ramSourceRam.add(b);

      return this;
    }

    public TraceBuilder ramSourceRom(final Boolean b) {
      if (filled.get(68)) {
        throw new IllegalStateException("RAM_SOURCE_ROM already set");
      } else {
        filled.set(68);
      }

      ramSourceRom.add(b);

      return this;
    }

    public TraceBuilder ramSourceStack(final Boolean b) {
      if (filled.get(69)) {
        throw new IllegalStateException("RAM_SOURCE_STACK already set");
      } else {
        filled.set(69);
      }

      ramSourceStack.add(b);

      return this;
    }

    public TraceBuilder ramSourceTxnData(final Boolean b) {
      if (filled.get(70)) {
        throw new IllegalStateException("RAM_SOURCE_TXN_DATA already set");
      } else {
        filled.set(70);
      }

      ramSourceTxnData.add(b);

      return this;
    }

    public TraceBuilder ramTargetBlakeData(final Boolean b) {
      if (filled.get(71)) {
        throw new IllegalStateException("RAM_TARGET_BLAKE_DATA already set");
      } else {
        filled.set(71);
      }

      ramTargetBlakeData.add(b);

      return this;
    }

    public TraceBuilder ramTargetEcData(final Boolean b) {
      if (filled.get(72)) {
        throw new IllegalStateException("RAM_TARGET_EC_DATA already set");
      } else {
        filled.set(72);
      }

      ramTargetEcData.add(b);

      return this;
    }

    public TraceBuilder ramTargetEcInfo(final Boolean b) {
      if (filled.get(73)) {
        throw new IllegalStateException("RAM_TARGET_EC_INFO already set");
      } else {
        filled.set(73);
      }

      ramTargetEcInfo.add(b);

      return this;
    }

    public TraceBuilder ramTargetHashData(final Boolean b) {
      if (filled.get(74)) {
        throw new IllegalStateException("RAM_TARGET_HASH_DATA already set");
      } else {
        filled.set(74);
      }

      ramTargetHashData.add(b);

      return this;
    }

    public TraceBuilder ramTargetHashInfo(final Boolean b) {
      if (filled.get(75)) {
        throw new IllegalStateException("RAM_TARGET_HASH_INFO already set");
      } else {
        filled.set(75);
      }

      ramTargetHashInfo.add(b);

      return this;
    }

    public TraceBuilder ramTargetLogData(final Boolean b) {
      if (filled.get(76)) {
        throw new IllegalStateException("RAM_TARGET_LOG_DATA already set");
      } else {
        filled.set(76);
      }

      ramTargetLogData.add(b);

      return this;
    }

    public TraceBuilder ramTargetModexpData(final Boolean b) {
      if (filled.get(77)) {
        throw new IllegalStateException("RAM_TARGET_MODEXP_DATA already set");
      } else {
        filled.set(77);
      }

      ramTargetModexpData.add(b);

      return this;
    }

    public TraceBuilder ramTargetRam(final Boolean b) {
      if (filled.get(78)) {
        throw new IllegalStateException("RAM_TARGET_RAM already set");
      } else {
        filled.set(78);
      }

      ramTargetRam.add(b);

      return this;
    }

    public TraceBuilder ramTargetRom(final Boolean b) {
      if (filled.get(79)) {
        throw new IllegalStateException("RAM_TARGET_ROM already set");
      } else {
        filled.set(79);
      }

      ramTargetRom.add(b);

      return this;
    }

    public TraceBuilder ramTargetStack(final Boolean b) {
      if (filled.get(80)) {
        throw new IllegalStateException("RAM_TARGET_STACK already set");
      } else {
        filled.set(80);
      }

      ramTargetStack.add(b);

      return this;
    }

    public TraceBuilder ramTargetTxnData(final Boolean b) {
      if (filled.get(81)) {
        throw new IllegalStateException("RAM_TARGET_TXN_DATA already set");
      } else {
        filled.set(81);
      }

      ramTargetTxnData.add(b);

      return this;
    }

    public TraceBuilder staticGas(final BigInteger b) {
      if (filled.get(82)) {
        throw new IllegalStateException("STATIC_GAS already set");
      } else {
        filled.set(82);
      }

      staticGas.add(b);

      return this;
    }

    public TraceBuilder twoLinesInstruction(final Boolean b) {
      if (filled.get(83)) {
        throw new IllegalStateException("TWO_LINES_INSTRUCTION already set");
      } else {
        filled.set(83);
      }

      twoLinesInstruction.add(b);

      return this;
    }

    public TraceBuilder validateRow() {
      if (!filled.get(0)) {
        throw new IllegalStateException("ADDRESS_TRIMMING_INSTRUCTION has not been filled");
      }

      if (!filled.get(1)) {
        throw new IllegalStateException("ALPHA has not been filled");
      }

      if (!filled.get(2)) {
        throw new IllegalStateException("BILLING_PER_BYTE has not been filled");
      }

      if (!filled.get(3)) {
        throw new IllegalStateException("BILLING_PER_WORD has not been filled");
      }

      if (!filled.get(4)) {
        throw new IllegalStateException("DELTA has not been filled");
      }

      if (!filled.get(5)) {
        throw new IllegalStateException("FAMILY_ACCOUNT has not been filled");
      }

      if (!filled.get(6)) {
        throw new IllegalStateException("FAMILY_ADD has not been filled");
      }

      if (!filled.get(7)) {
        throw new IllegalStateException("FAMILY_BATCH has not been filled");
      }

      if (!filled.get(8)) {
        throw new IllegalStateException("FAMILY_BIN has not been filled");
      }

      if (!filled.get(9)) {
        throw new IllegalStateException("FAMILY_CALL has not been filled");
      }

      if (!filled.get(10)) {
        throw new IllegalStateException("FAMILY_CONTEXT has not been filled");
      }

      if (!filled.get(11)) {
        throw new IllegalStateException("FAMILY_COPY has not been filled");
      }

      if (!filled.get(12)) {
        throw new IllegalStateException("FAMILY_CREATE has not been filled");
      }

      if (!filled.get(13)) {
        throw new IllegalStateException("FAMILY_DUP has not been filled");
      }

      if (!filled.get(14)) {
        throw new IllegalStateException("FAMILY_EXT has not been filled");
      }

      if (!filled.get(15)) {
        throw new IllegalStateException("FAMILY_HALT has not been filled");
      }

      if (!filled.get(16)) {
        throw new IllegalStateException("FAMILY_INVALID has not been filled");
      }

      if (!filled.get(17)) {
        throw new IllegalStateException("FAMILY_JUMP has not been filled");
      }

      if (!filled.get(18)) {
        throw new IllegalStateException("FAMILY_KEC has not been filled");
      }

      if (!filled.get(19)) {
        throw new IllegalStateException("FAMILY_LOG has not been filled");
      }

      if (!filled.get(20)) {
        throw new IllegalStateException("FAMILY_MACHINE_STATE has not been filled");
      }

      if (!filled.get(21)) {
        throw new IllegalStateException("FAMILY_MOD has not been filled");
      }

      if (!filled.get(22)) {
        throw new IllegalStateException("FAMILY_MUL has not been filled");
      }

      if (!filled.get(23)) {
        throw new IllegalStateException("FAMILY_PUSH_POP has not been filled");
      }

      if (!filled.get(24)) {
        throw new IllegalStateException("FAMILY_SHF has not been filled");
      }

      if (!filled.get(25)) {
        throw new IllegalStateException("FAMILY_STACK_RAM has not been filled");
      }

      if (!filled.get(26)) {
        throw new IllegalStateException("FAMILY_STORAGE has not been filled");
      }

      if (!filled.get(27)) {
        throw new IllegalStateException("FAMILY_SWAP has not been filled");
      }

      if (!filled.get(28)) {
        throw new IllegalStateException("FAMILY_TRANSACTION has not been filled");
      }

      if (!filled.get(29)) {
        throw new IllegalStateException("FAMILY_WCP has not been filled");
      }

      if (!filled.get(30)) {
        throw new IllegalStateException("FLAG1 has not been filled");
      }

      if (!filled.get(31)) {
        throw new IllegalStateException("FLAG2 has not been filled");
      }

      if (!filled.get(32)) {
        throw new IllegalStateException("FLAG3 has not been filled");
      }

      if (!filled.get(33)) {
        throw new IllegalStateException("FLAG4 has not been filled");
      }

      if (!filled.get(34)) {
        throw new IllegalStateException("FORBIDDEN_IN_STATIC has not been filled");
      }

      if (!filled.get(35)) {
        throw new IllegalStateException("IS_JUMPDEST has not been filled");
      }

      if (!filled.get(36)) {
        throw new IllegalStateException("IS_PUSH has not been filled");
      }

      if (!filled.get(37)) {
        throw new IllegalStateException("MXP_TYPE_1 has not been filled");
      }

      if (!filled.get(38)) {
        throw new IllegalStateException("MXP_TYPE_2 has not been filled");
      }

      if (!filled.get(39)) {
        throw new IllegalStateException("MXP_TYPE_3 has not been filled");
      }

      if (!filled.get(40)) {
        throw new IllegalStateException("MXP_TYPE_4 has not been filled");
      }

      if (!filled.get(41)) {
        throw new IllegalStateException("MXP_TYPE_5 has not been filled");
      }

      if (!filled.get(42)) {
        throw new IllegalStateException("NB_ADDED has not been filled");
      }

      if (!filled.get(43)) {
        throw new IllegalStateException("NB_REMOVED has not been filled");
      }

      if (!filled.get(44)) {
        throw new IllegalStateException("OPCODE has not been filled");
      }

      if (!filled.get(45)) {
        throw new IllegalStateException("PATTERN_CALL has not been filled");
      }

      if (!filled.get(46)) {
        throw new IllegalStateException("PATTERN_COPY has not been filled");
      }

      if (!filled.get(47)) {
        throw new IllegalStateException("PATTERN_CREATE has not been filled");
      }

      if (!filled.get(48)) {
        throw new IllegalStateException("PATTERN_DUP has not been filled");
      }

      if (!filled.get(49)) {
        throw new IllegalStateException("PATTERN_LOAD_STORE has not been filled");
      }

      if (!filled.get(50)) {
        throw new IllegalStateException("PATTERN_LOG has not been filled");
      }

      if (!filled.get(51)) {
        throw new IllegalStateException("PATTERN_ONE_ONE has not been filled");
      }

      if (!filled.get(52)) {
        throw new IllegalStateException("PATTERN_ONE_ZERO has not been filled");
      }

      if (!filled.get(53)) {
        throw new IllegalStateException("PATTERN_SWAP has not been filled");
      }

      if (!filled.get(54)) {
        throw new IllegalStateException("PATTERN_THREE_ONE has not been filled");
      }

      if (!filled.get(55)) {
        throw new IllegalStateException("PATTERN_TWO_ONE has not been filled");
      }

      if (!filled.get(56)) {
        throw new IllegalStateException("PATTERN_TWO_ZERO has not been filled");
      }

      if (!filled.get(57)) {
        throw new IllegalStateException("PATTERN_ZERO_ONE has not been filled");
      }

      if (!filled.get(58)) {
        throw new IllegalStateException("PATTERN_ZERO_ZERO has not been filled");
      }

      if (!filled.get(59)) {
        throw new IllegalStateException("RAM_ENABLED has not been filled");
      }

      if (!filled.get(60)) {
        throw new IllegalStateException("RAM_SOURCE_BLAKE_DATA has not been filled");
      }

      if (!filled.get(61)) {
        throw new IllegalStateException("RAM_SOURCE_EC_DATA has not been filled");
      }

      if (!filled.get(62)) {
        throw new IllegalStateException("RAM_SOURCE_EC_INFO has not been filled");
      }

      if (!filled.get(63)) {
        throw new IllegalStateException("RAM_SOURCE_HASH_DATA has not been filled");
      }

      if (!filled.get(64)) {
        throw new IllegalStateException("RAM_SOURCE_HASH_INFO has not been filled");
      }

      if (!filled.get(65)) {
        throw new IllegalStateException("RAM_SOURCE_LOG_DATA has not been filled");
      }

      if (!filled.get(66)) {
        throw new IllegalStateException("RAM_SOURCE_MODEXP_DATA has not been filled");
      }

      if (!filled.get(67)) {
        throw new IllegalStateException("RAM_SOURCE_RAM has not been filled");
      }

      if (!filled.get(68)) {
        throw new IllegalStateException("RAM_SOURCE_ROM has not been filled");
      }

      if (!filled.get(69)) {
        throw new IllegalStateException("RAM_SOURCE_STACK has not been filled");
      }

      if (!filled.get(70)) {
        throw new IllegalStateException("RAM_SOURCE_TXN_DATA has not been filled");
      }

      if (!filled.get(71)) {
        throw new IllegalStateException("RAM_TARGET_BLAKE_DATA has not been filled");
      }

      if (!filled.get(72)) {
        throw new IllegalStateException("RAM_TARGET_EC_DATA has not been filled");
      }

      if (!filled.get(73)) {
        throw new IllegalStateException("RAM_TARGET_EC_INFO has not been filled");
      }

      if (!filled.get(74)) {
        throw new IllegalStateException("RAM_TARGET_HASH_DATA has not been filled");
      }

      if (!filled.get(75)) {
        throw new IllegalStateException("RAM_TARGET_HASH_INFO has not been filled");
      }

      if (!filled.get(76)) {
        throw new IllegalStateException("RAM_TARGET_LOG_DATA has not been filled");
      }

      if (!filled.get(77)) {
        throw new IllegalStateException("RAM_TARGET_MODEXP_DATA has not been filled");
      }

      if (!filled.get(78)) {
        throw new IllegalStateException("RAM_TARGET_RAM has not been filled");
      }

      if (!filled.get(79)) {
        throw new IllegalStateException("RAM_TARGET_ROM has not been filled");
      }

      if (!filled.get(80)) {
        throw new IllegalStateException("RAM_TARGET_STACK has not been filled");
      }

      if (!filled.get(81)) {
        throw new IllegalStateException("RAM_TARGET_TXN_DATA has not been filled");
      }

      if (!filled.get(82)) {
        throw new IllegalStateException("STATIC_GAS has not been filled");
      }

      if (!filled.get(83)) {
        throw new IllegalStateException("TWO_LINES_INSTRUCTION has not been filled");
      }

      filled.clear();

      return this;
    }

    public TraceBuilder fillAndValidateRow() {
      if (!filled.get(0)) {
        addressTrimmingInstruction.add(false);
        this.filled.set(0);
      }
      if (!filled.get(1)) {
        alpha.add(UnsignedByte.of(0));
        this.filled.set(1);
      }
      if (!filled.get(2)) {
        billingPerByte.add(BigInteger.ZERO);
        this.filled.set(2);
      }
      if (!filled.get(3)) {
        billingPerWord.add(BigInteger.ZERO);
        this.filled.set(3);
      }
      if (!filled.get(4)) {
        delta.add(UnsignedByte.of(0));
        this.filled.set(4);
      }
      if (!filled.get(5)) {
        familyAccount.add(false);
        this.filled.set(5);
      }
      if (!filled.get(6)) {
        familyAdd.add(false);
        this.filled.set(6);
      }
      if (!filled.get(7)) {
        familyBatch.add(false);
        this.filled.set(7);
      }
      if (!filled.get(8)) {
        familyBin.add(false);
        this.filled.set(8);
      }
      if (!filled.get(9)) {
        familyCall.add(false);
        this.filled.set(9);
      }
      if (!filled.get(10)) {
        familyContext.add(false);
        this.filled.set(10);
      }
      if (!filled.get(11)) {
        familyCopy.add(false);
        this.filled.set(11);
      }
      if (!filled.get(12)) {
        familyCreate.add(false);
        this.filled.set(12);
      }
      if (!filled.get(13)) {
        familyDup.add(false);
        this.filled.set(13);
      }
      if (!filled.get(14)) {
        familyExt.add(false);
        this.filled.set(14);
      }
      if (!filled.get(15)) {
        familyHalt.add(false);
        this.filled.set(15);
      }
      if (!filled.get(16)) {
        familyInvalid.add(false);
        this.filled.set(16);
      }
      if (!filled.get(17)) {
        familyJump.add(false);
        this.filled.set(17);
      }
      if (!filled.get(18)) {
        familyKec.add(false);
        this.filled.set(18);
      }
      if (!filled.get(19)) {
        familyLog.add(false);
        this.filled.set(19);
      }
      if (!filled.get(20)) {
        familyMachineState.add(false);
        this.filled.set(20);
      }
      if (!filled.get(21)) {
        familyMod.add(false);
        this.filled.set(21);
      }
      if (!filled.get(22)) {
        familyMul.add(false);
        this.filled.set(22);
      }
      if (!filled.get(23)) {
        familyPushPop.add(false);
        this.filled.set(23);
      }
      if (!filled.get(24)) {
        familyShf.add(false);
        this.filled.set(24);
      }
      if (!filled.get(25)) {
        familyStackRam.add(false);
        this.filled.set(25);
      }
      if (!filled.get(26)) {
        familyStorage.add(false);
        this.filled.set(26);
      }
      if (!filled.get(27)) {
        familySwap.add(false);
        this.filled.set(27);
      }
      if (!filled.get(28)) {
        familyTransaction.add(false);
        this.filled.set(28);
      }
      if (!filled.get(29)) {
        familyWcp.add(false);
        this.filled.set(29);
      }
      if (!filled.get(30)) {
        flag1.add(false);
        this.filled.set(30);
      }
      if (!filled.get(31)) {
        flag2.add(false);
        this.filled.set(31);
      }
      if (!filled.get(32)) {
        flag3.add(false);
        this.filled.set(32);
      }
      if (!filled.get(33)) {
        flag4.add(false);
        this.filled.set(33);
      }
      if (!filled.get(34)) {
        forbiddenInStatic.add(false);
        this.filled.set(34);
      }
      if (!filled.get(35)) {
        isJumpdest.add(false);
        this.filled.set(35);
      }
      if (!filled.get(36)) {
        isPush.add(false);
        this.filled.set(36);
      }
      if (!filled.get(37)) {
        mxpType1.add(false);
        this.filled.set(37);
      }
      if (!filled.get(38)) {
        mxpType2.add(false);
        this.filled.set(38);
      }
      if (!filled.get(39)) {
        mxpType3.add(false);
        this.filled.set(39);
      }
      if (!filled.get(40)) {
        mxpType4.add(false);
        this.filled.set(40);
      }
      if (!filled.get(41)) {
        mxpType5.add(false);
        this.filled.set(41);
      }
      if (!filled.get(42)) {
        nbAdded.add(UnsignedByte.of(0));
        this.filled.set(42);
      }
      if (!filled.get(43)) {
        nbRemoved.add(UnsignedByte.of(0));
        this.filled.set(43);
      }
      if (!filled.get(44)) {
        opcode.add(BigInteger.ZERO);
        this.filled.set(44);
      }
      if (!filled.get(45)) {
        patternCall.add(false);
        this.filled.set(45);
      }
      if (!filled.get(46)) {
        patternCopy.add(false);
        this.filled.set(46);
      }
      if (!filled.get(47)) {
        patternCreate.add(false);
        this.filled.set(47);
      }
      if (!filled.get(48)) {
        patternDup.add(false);
        this.filled.set(48);
      }
      if (!filled.get(49)) {
        patternLoadStore.add(false);
        this.filled.set(49);
      }
      if (!filled.get(50)) {
        patternLog.add(false);
        this.filled.set(50);
      }
      if (!filled.get(51)) {
        patternOneOne.add(false);
        this.filled.set(51);
      }
      if (!filled.get(52)) {
        patternOneZero.add(false);
        this.filled.set(52);
      }
      if (!filled.get(53)) {
        patternSwap.add(false);
        this.filled.set(53);
      }
      if (!filled.get(54)) {
        patternThreeOne.add(false);
        this.filled.set(54);
      }
      if (!filled.get(55)) {
        patternTwoOne.add(false);
        this.filled.set(55);
      }
      if (!filled.get(56)) {
        patternTwoZero.add(false);
        this.filled.set(56);
      }
      if (!filled.get(57)) {
        patternZeroOne.add(false);
        this.filled.set(57);
      }
      if (!filled.get(58)) {
        patternZeroZero.add(false);
        this.filled.set(58);
      }
      if (!filled.get(59)) {
        ramEnabled.add(false);
        this.filled.set(59);
      }
      if (!filled.get(60)) {
        ramSourceBlakeData.add(false);
        this.filled.set(60);
      }
      if (!filled.get(61)) {
        ramSourceEcData.add(false);
        this.filled.set(61);
      }
      if (!filled.get(62)) {
        ramSourceEcInfo.add(false);
        this.filled.set(62);
      }
      if (!filled.get(63)) {
        ramSourceHashData.add(false);
        this.filled.set(63);
      }
      if (!filled.get(64)) {
        ramSourceHashInfo.add(false);
        this.filled.set(64);
      }
      if (!filled.get(65)) {
        ramSourceLogData.add(false);
        this.filled.set(65);
      }
      if (!filled.get(66)) {
        ramSourceModexpData.add(false);
        this.filled.set(66);
      }
      if (!filled.get(67)) {
        ramSourceRam.add(false);
        this.filled.set(67);
      }
      if (!filled.get(68)) {
        ramSourceRom.add(false);
        this.filled.set(68);
      }
      if (!filled.get(69)) {
        ramSourceStack.add(false);
        this.filled.set(69);
      }
      if (!filled.get(70)) {
        ramSourceTxnData.add(false);
        this.filled.set(70);
      }
      if (!filled.get(71)) {
        ramTargetBlakeData.add(false);
        this.filled.set(71);
      }
      if (!filled.get(72)) {
        ramTargetEcData.add(false);
        this.filled.set(72);
      }
      if (!filled.get(73)) {
        ramTargetEcInfo.add(false);
        this.filled.set(73);
      }
      if (!filled.get(74)) {
        ramTargetHashData.add(false);
        this.filled.set(74);
      }
      if (!filled.get(75)) {
        ramTargetHashInfo.add(false);
        this.filled.set(75);
      }
      if (!filled.get(76)) {
        ramTargetLogData.add(false);
        this.filled.set(76);
      }
      if (!filled.get(77)) {
        ramTargetModexpData.add(false);
        this.filled.set(77);
      }
      if (!filled.get(78)) {
        ramTargetRam.add(false);
        this.filled.set(78);
      }
      if (!filled.get(79)) {
        ramTargetRom.add(false);
        this.filled.set(79);
      }
      if (!filled.get(80)) {
        ramTargetStack.add(false);
        this.filled.set(80);
      }
      if (!filled.get(81)) {
        ramTargetTxnData.add(false);
        this.filled.set(81);
      }
      if (!filled.get(82)) {
        staticGas.add(BigInteger.ZERO);
        this.filled.set(82);
      }
      if (!filled.get(83)) {
        twoLinesInstruction.add(false);
        this.filled.set(83);
      }

      return this.validateRow();
    }

    public Trace build() {
      if (!filled.isEmpty()) {
        throw new IllegalStateException("Cannot build trace with a non-validated row.");
      }

      return new Trace(
          addressTrimmingInstruction,
          alpha,
          billingPerByte,
          billingPerWord,
          delta,
          familyAccount,
          familyAdd,
          familyBatch,
          familyBin,
          familyCall,
          familyContext,
          familyCopy,
          familyCreate,
          familyDup,
          familyExt,
          familyHalt,
          familyInvalid,
          familyJump,
          familyKec,
          familyLog,
          familyMachineState,
          familyMod,
          familyMul,
          familyPushPop,
          familyShf,
          familyStackRam,
          familyStorage,
          familySwap,
          familyTransaction,
          familyWcp,
          flag1,
          flag2,
          flag3,
          flag4,
          forbiddenInStatic,
          isJumpdest,
          isPush,
          mxpType1,
          mxpType2,
          mxpType3,
          mxpType4,
          mxpType5,
          nbAdded,
          nbRemoved,
          opcode,
          patternCall,
          patternCopy,
          patternCreate,
          patternDup,
          patternLoadStore,
          patternLog,
          patternOneOne,
          patternOneZero,
          patternSwap,
          patternThreeOne,
          patternTwoOne,
          patternTwoZero,
          patternZeroOne,
          patternZeroZero,
          ramEnabled,
          ramSourceBlakeData,
          ramSourceEcData,
          ramSourceEcInfo,
          ramSourceHashData,
          ramSourceHashInfo,
          ramSourceLogData,
          ramSourceModexpData,
          ramSourceRam,
          ramSourceRom,
          ramSourceStack,
          ramSourceTxnData,
          ramTargetBlakeData,
          ramTargetEcData,
          ramTargetEcInfo,
          ramTargetHashData,
          ramTargetHashInfo,
          ramTargetLogData,
          ramTargetModexpData,
          ramTargetRam,
          ramTargetRom,
          ramTargetStack,
          ramTargetTxnData,
          staticGas,
          twoLinesInstruction);
    }
  }
}
