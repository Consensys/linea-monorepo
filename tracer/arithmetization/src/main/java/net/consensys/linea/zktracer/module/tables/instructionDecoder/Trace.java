/*
 * Copyright ConsenSys Inc.
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

import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

/**
 * WARNING: This code is generated automatically.
 *
 * <p>Any modifications to this code may be overwritten and could lead to unexpected behavior.
 * Please DO NOT ATTEMPT TO MODIFY this code directly.
 */
public class Trace {

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer addressTrimmingInstruction;
  private final MappedByteBuffer alpha;
  private final MappedByteBuffer billingPerByte;
  private final MappedByteBuffer billingPerWord;
  private final MappedByteBuffer delta;
  private final MappedByteBuffer familyAccount;
  private final MappedByteBuffer familyAdd;
  private final MappedByteBuffer familyBatch;
  private final MappedByteBuffer familyBin;
  private final MappedByteBuffer familyCall;
  private final MappedByteBuffer familyContext;
  private final MappedByteBuffer familyCopy;
  private final MappedByteBuffer familyCreate;
  private final MappedByteBuffer familyDup;
  private final MappedByteBuffer familyExt;
  private final MappedByteBuffer familyHalt;
  private final MappedByteBuffer familyInvalid;
  private final MappedByteBuffer familyJump;
  private final MappedByteBuffer familyKec;
  private final MappedByteBuffer familyLog;
  private final MappedByteBuffer familyMachineState;
  private final MappedByteBuffer familyMod;
  private final MappedByteBuffer familyMul;
  private final MappedByteBuffer familyPushPop;
  private final MappedByteBuffer familyShf;
  private final MappedByteBuffer familyStackRam;
  private final MappedByteBuffer familyStorage;
  private final MappedByteBuffer familySwap;
  private final MappedByteBuffer familyTransaction;
  private final MappedByteBuffer familyWcp;
  private final MappedByteBuffer flag1;
  private final MappedByteBuffer flag2;
  private final MappedByteBuffer flag3;
  private final MappedByteBuffer flag4;
  private final MappedByteBuffer isJumpdest;
  private final MappedByteBuffer isPush;
  private final MappedByteBuffer mxpFlag;
  private final MappedByteBuffer mxpType1;
  private final MappedByteBuffer mxpType2;
  private final MappedByteBuffer mxpType3;
  private final MappedByteBuffer mxpType4;
  private final MappedByteBuffer mxpType5;
  private final MappedByteBuffer nbAdded;
  private final MappedByteBuffer nbRemoved;
  private final MappedByteBuffer opcode;
  private final MappedByteBuffer patternCall;
  private final MappedByteBuffer patternCopy;
  private final MappedByteBuffer patternCreate;
  private final MappedByteBuffer patternDup;
  private final MappedByteBuffer patternLoadStore;
  private final MappedByteBuffer patternLog;
  private final MappedByteBuffer patternOneOne;
  private final MappedByteBuffer patternOneZero;
  private final MappedByteBuffer patternSwap;
  private final MappedByteBuffer patternThreeOne;
  private final MappedByteBuffer patternTwoOne;
  private final MappedByteBuffer patternTwoZero;
  private final MappedByteBuffer patternZeroOne;
  private final MappedByteBuffer patternZeroZero;
  private final MappedByteBuffer ramEnabled;
  private final MappedByteBuffer ramSourceBlakeData;
  private final MappedByteBuffer ramSourceEcData;
  private final MappedByteBuffer ramSourceEcInfo;
  private final MappedByteBuffer ramSourceHashData;
  private final MappedByteBuffer ramSourceHashInfo;
  private final MappedByteBuffer ramSourceLogData;
  private final MappedByteBuffer ramSourceModexpData;
  private final MappedByteBuffer ramSourceRam;
  private final MappedByteBuffer ramSourceRom;
  private final MappedByteBuffer ramSourceStack;
  private final MappedByteBuffer ramSourceTxnData;
  private final MappedByteBuffer ramTargetBlakeData;
  private final MappedByteBuffer ramTargetEcData;
  private final MappedByteBuffer ramTargetEcInfo;
  private final MappedByteBuffer ramTargetHashData;
  private final MappedByteBuffer ramTargetHashInfo;
  private final MappedByteBuffer ramTargetLogData;
  private final MappedByteBuffer ramTargetModexpData;
  private final MappedByteBuffer ramTargetRam;
  private final MappedByteBuffer ramTargetRom;
  private final MappedByteBuffer ramTargetStack;
  private final MappedByteBuffer ramTargetTxnData;
  private final MappedByteBuffer staticFlag;
  private final MappedByteBuffer staticGas;
  private final MappedByteBuffer twoLineInstruction;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("instdecoder.ADDRESS_TRIMMING_INSTRUCTION", 1, length),
        new ColumnHeader("instdecoder.ALPHA", 1, length),
        new ColumnHeader("instdecoder.BILLING_PER_BYTE", 32, length),
        new ColumnHeader("instdecoder.BILLING_PER_WORD", 32, length),
        new ColumnHeader("instdecoder.DELTA", 1, length),
        new ColumnHeader("instdecoder.FAMILY_ACCOUNT", 1, length),
        new ColumnHeader("instdecoder.FAMILY_ADD", 1, length),
        new ColumnHeader("instdecoder.FAMILY_BATCH", 1, length),
        new ColumnHeader("instdecoder.FAMILY_BIN", 1, length),
        new ColumnHeader("instdecoder.FAMILY_CALL", 1, length),
        new ColumnHeader("instdecoder.FAMILY_CONTEXT", 1, length),
        new ColumnHeader("instdecoder.FAMILY_COPY", 1, length),
        new ColumnHeader("instdecoder.FAMILY_CREATE", 1, length),
        new ColumnHeader("instdecoder.FAMILY_DUP", 1, length),
        new ColumnHeader("instdecoder.FAMILY_EXT", 1, length),
        new ColumnHeader("instdecoder.FAMILY_HALT", 1, length),
        new ColumnHeader("instdecoder.FAMILY_INVALID", 1, length),
        new ColumnHeader("instdecoder.FAMILY_JUMP", 1, length),
        new ColumnHeader("instdecoder.FAMILY_KEC", 1, length),
        new ColumnHeader("instdecoder.FAMILY_LOG", 1, length),
        new ColumnHeader("instdecoder.FAMILY_MACHINE_STATE", 1, length),
        new ColumnHeader("instdecoder.FAMILY_MOD", 1, length),
        new ColumnHeader("instdecoder.FAMILY_MUL", 1, length),
        new ColumnHeader("instdecoder.FAMILY_PUSH_POP", 1, length),
        new ColumnHeader("instdecoder.FAMILY_SHF", 1, length),
        new ColumnHeader("instdecoder.FAMILY_STACK_RAM", 1, length),
        new ColumnHeader("instdecoder.FAMILY_STORAGE", 1, length),
        new ColumnHeader("instdecoder.FAMILY_SWAP", 1, length),
        new ColumnHeader("instdecoder.FAMILY_TRANSACTION", 1, length),
        new ColumnHeader("instdecoder.FAMILY_WCP", 1, length),
        new ColumnHeader("instdecoder.FLAG_1", 1, length),
        new ColumnHeader("instdecoder.FLAG_2", 1, length),
        new ColumnHeader("instdecoder.FLAG_3", 1, length),
        new ColumnHeader("instdecoder.FLAG_4", 1, length),
        new ColumnHeader("instdecoder.IS_JUMPDEST", 1, length),
        new ColumnHeader("instdecoder.IS_PUSH", 1, length),
        new ColumnHeader("instdecoder.MXP_FLAG", 1, length),
        new ColumnHeader("instdecoder.MXP_TYPE_1", 1, length),
        new ColumnHeader("instdecoder.MXP_TYPE_2", 1, length),
        new ColumnHeader("instdecoder.MXP_TYPE_3", 1, length),
        new ColumnHeader("instdecoder.MXP_TYPE_4", 1, length),
        new ColumnHeader("instdecoder.MXP_TYPE_5", 1, length),
        new ColumnHeader("instdecoder.NB_ADDED", 1, length),
        new ColumnHeader("instdecoder.NB_REMOVED", 1, length),
        new ColumnHeader("instdecoder.OPCODE", 32, length),
        new ColumnHeader("instdecoder.PATTERN_CALL", 1, length),
        new ColumnHeader("instdecoder.PATTERN_COPY", 1, length),
        new ColumnHeader("instdecoder.PATTERN_CREATE", 1, length),
        new ColumnHeader("instdecoder.PATTERN_DUP", 1, length),
        new ColumnHeader("instdecoder.PATTERN_LOAD_STORE", 1, length),
        new ColumnHeader("instdecoder.PATTERN_LOG", 1, length),
        new ColumnHeader("instdecoder.PATTERN_ONE_ONE", 1, length),
        new ColumnHeader("instdecoder.PATTERN_ONE_ZERO", 1, length),
        new ColumnHeader("instdecoder.PATTERN_SWAP", 1, length),
        new ColumnHeader("instdecoder.PATTERN_THREE_ONE", 1, length),
        new ColumnHeader("instdecoder.PATTERN_TWO_ONE", 1, length),
        new ColumnHeader("instdecoder.PATTERN_TWO_ZERO", 1, length),
        new ColumnHeader("instdecoder.PATTERN_ZERO_ONE", 1, length),
        new ColumnHeader("instdecoder.PATTERN_ZERO_ZERO", 1, length),
        new ColumnHeader("instdecoder.RAM_ENABLED", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_BLAKE_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_EC_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_EC_INFO", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_HASH_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_HASH_INFO", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_LOG_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_MODEXP_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_RAM", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_ROM", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_STACK", 1, length),
        new ColumnHeader("instdecoder.RAM_SOURCE_TXN_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_BLAKE_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_EC_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_EC_INFO", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_HASH_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_HASH_INFO", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_LOG_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_MODEXP_DATA", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_RAM", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_ROM", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_STACK", 1, length),
        new ColumnHeader("instdecoder.RAM_TARGET_TXN_DATA", 1, length),
        new ColumnHeader("instdecoder.STATIC_FLAG", 1, length),
        new ColumnHeader("instdecoder.STATIC_GAS", 32, length),
        new ColumnHeader("instdecoder.TWO_LINE_INSTRUCTION", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.addressTrimmingInstruction = buffers.get(0);
    this.alpha = buffers.get(1);
    this.billingPerByte = buffers.get(2);
    this.billingPerWord = buffers.get(3);
    this.delta = buffers.get(4);
    this.familyAccount = buffers.get(5);
    this.familyAdd = buffers.get(6);
    this.familyBatch = buffers.get(7);
    this.familyBin = buffers.get(8);
    this.familyCall = buffers.get(9);
    this.familyContext = buffers.get(10);
    this.familyCopy = buffers.get(11);
    this.familyCreate = buffers.get(12);
    this.familyDup = buffers.get(13);
    this.familyExt = buffers.get(14);
    this.familyHalt = buffers.get(15);
    this.familyInvalid = buffers.get(16);
    this.familyJump = buffers.get(17);
    this.familyKec = buffers.get(18);
    this.familyLog = buffers.get(19);
    this.familyMachineState = buffers.get(20);
    this.familyMod = buffers.get(21);
    this.familyMul = buffers.get(22);
    this.familyPushPop = buffers.get(23);
    this.familyShf = buffers.get(24);
    this.familyStackRam = buffers.get(25);
    this.familyStorage = buffers.get(26);
    this.familySwap = buffers.get(27);
    this.familyTransaction = buffers.get(28);
    this.familyWcp = buffers.get(29);
    this.flag1 = buffers.get(30);
    this.flag2 = buffers.get(31);
    this.flag3 = buffers.get(32);
    this.flag4 = buffers.get(33);
    this.isJumpdest = buffers.get(34);
    this.isPush = buffers.get(35);
    this.mxpFlag = buffers.get(36);
    this.mxpType1 = buffers.get(37);
    this.mxpType2 = buffers.get(38);
    this.mxpType3 = buffers.get(39);
    this.mxpType4 = buffers.get(40);
    this.mxpType5 = buffers.get(41);
    this.nbAdded = buffers.get(42);
    this.nbRemoved = buffers.get(43);
    this.opcode = buffers.get(44);
    this.patternCall = buffers.get(45);
    this.patternCopy = buffers.get(46);
    this.patternCreate = buffers.get(47);
    this.patternDup = buffers.get(48);
    this.patternLoadStore = buffers.get(49);
    this.patternLog = buffers.get(50);
    this.patternOneOne = buffers.get(51);
    this.patternOneZero = buffers.get(52);
    this.patternSwap = buffers.get(53);
    this.patternThreeOne = buffers.get(54);
    this.patternTwoOne = buffers.get(55);
    this.patternTwoZero = buffers.get(56);
    this.patternZeroOne = buffers.get(57);
    this.patternZeroZero = buffers.get(58);
    this.ramEnabled = buffers.get(59);
    this.ramSourceBlakeData = buffers.get(60);
    this.ramSourceEcData = buffers.get(61);
    this.ramSourceEcInfo = buffers.get(62);
    this.ramSourceHashData = buffers.get(63);
    this.ramSourceHashInfo = buffers.get(64);
    this.ramSourceLogData = buffers.get(65);
    this.ramSourceModexpData = buffers.get(66);
    this.ramSourceRam = buffers.get(67);
    this.ramSourceRom = buffers.get(68);
    this.ramSourceStack = buffers.get(69);
    this.ramSourceTxnData = buffers.get(70);
    this.ramTargetBlakeData = buffers.get(71);
    this.ramTargetEcData = buffers.get(72);
    this.ramTargetEcInfo = buffers.get(73);
    this.ramTargetHashData = buffers.get(74);
    this.ramTargetHashInfo = buffers.get(75);
    this.ramTargetLogData = buffers.get(76);
    this.ramTargetModexpData = buffers.get(77);
    this.ramTargetRam = buffers.get(78);
    this.ramTargetRom = buffers.get(79);
    this.ramTargetStack = buffers.get(80);
    this.ramTargetTxnData = buffers.get(81);
    this.staticFlag = buffers.get(82);
    this.staticGas = buffers.get(83);
    this.twoLineInstruction = buffers.get(84);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace addressTrimmingInstruction(final Boolean b) {
    if (filled.get(0)) {
      throw new IllegalStateException("instdecoder.ADDRESS_TRIMMING_INSTRUCTION already set");
    } else {
      filled.set(0);
    }

    addressTrimmingInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace alpha(final UnsignedByte b) {
    if (filled.get(1)) {
      throw new IllegalStateException("instdecoder.ALPHA already set");
    } else {
      filled.set(1);
    }

    alpha.put(b.toByte());

    return this;
  }

  public Trace billingPerByte(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_BYTE already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      billingPerByte.put((byte) 0);
    }
    billingPerByte.put(b.toArrayUnsafe());

    return this;
  }

  public Trace billingPerWord(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_WORD already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      billingPerWord.put((byte) 0);
    }
    billingPerWord.put(b.toArrayUnsafe());

    return this;
  }

  public Trace delta(final UnsignedByte b) {
    if (filled.get(4)) {
      throw new IllegalStateException("instdecoder.DELTA already set");
    } else {
      filled.set(4);
    }

    delta.put(b.toByte());

    return this;
  }

  public Trace familyAccount(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("instdecoder.FAMILY_ACCOUNT already set");
    } else {
      filled.set(5);
    }

    familyAccount.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyAdd(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("instdecoder.FAMILY_ADD already set");
    } else {
      filled.set(6);
    }

    familyAdd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyBatch(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("instdecoder.FAMILY_BATCH already set");
    } else {
      filled.set(7);
    }

    familyBatch.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyBin(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("instdecoder.FAMILY_BIN already set");
    } else {
      filled.set(8);
    }

    familyBin.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyCall(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("instdecoder.FAMILY_CALL already set");
    } else {
      filled.set(9);
    }

    familyCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyContext(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("instdecoder.FAMILY_CONTEXT already set");
    } else {
      filled.set(10);
    }

    familyContext.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyCopy(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("instdecoder.FAMILY_COPY already set");
    } else {
      filled.set(11);
    }

    familyCopy.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyCreate(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("instdecoder.FAMILY_CREATE already set");
    } else {
      filled.set(12);
    }

    familyCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyDup(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("instdecoder.FAMILY_DUP already set");
    } else {
      filled.set(13);
    }

    familyDup.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyExt(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("instdecoder.FAMILY_EXT already set");
    } else {
      filled.set(14);
    }

    familyExt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyHalt(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("instdecoder.FAMILY_HALT already set");
    } else {
      filled.set(15);
    }

    familyHalt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyInvalid(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("instdecoder.FAMILY_INVALID already set");
    } else {
      filled.set(16);
    }

    familyInvalid.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyJump(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("instdecoder.FAMILY_JUMP already set");
    } else {
      filled.set(17);
    }

    familyJump.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyKec(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("instdecoder.FAMILY_KEC already set");
    } else {
      filled.set(18);
    }

    familyKec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyLog(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("instdecoder.FAMILY_LOG already set");
    } else {
      filled.set(19);
    }

    familyLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyMachineState(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("instdecoder.FAMILY_MACHINE_STATE already set");
    } else {
      filled.set(20);
    }

    familyMachineState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyMod(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("instdecoder.FAMILY_MOD already set");
    } else {
      filled.set(21);
    }

    familyMod.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyMul(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("instdecoder.FAMILY_MUL already set");
    } else {
      filled.set(22);
    }

    familyMul.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyPushPop(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("instdecoder.FAMILY_PUSH_POP already set");
    } else {
      filled.set(23);
    }

    familyPushPop.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyShf(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("instdecoder.FAMILY_SHF already set");
    } else {
      filled.set(24);
    }

    familyShf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyStackRam(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("instdecoder.FAMILY_STACK_RAM already set");
    } else {
      filled.set(25);
    }

    familyStackRam.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyStorage(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("instdecoder.FAMILY_STORAGE already set");
    } else {
      filled.set(26);
    }

    familyStorage.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familySwap(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("instdecoder.FAMILY_SWAP already set");
    } else {
      filled.set(27);
    }

    familySwap.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyTransaction(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("instdecoder.FAMILY_TRANSACTION already set");
    } else {
      filled.set(28);
    }

    familyTransaction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyWcp(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("instdecoder.FAMILY_WCP already set");
    } else {
      filled.set(29);
    }

    familyWcp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag1(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("instdecoder.FLAG_1 already set");
    } else {
      filled.set(30);
    }

    flag1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag2(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("instdecoder.FLAG_2 already set");
    } else {
      filled.set(31);
    }

    flag2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag3(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("instdecoder.FLAG_3 already set");
    } else {
      filled.set(32);
    }

    flag3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag4(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("instdecoder.FLAG_4 already set");
    } else {
      filled.set(33);
    }

    flag4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJumpdest(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("instdecoder.IS_JUMPDEST already set");
    } else {
      filled.set(34);
    }

    isJumpdest.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPush(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("instdecoder.IS_PUSH already set");
    } else {
      filled.set(35);
    }

    isPush.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpFlag(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("instdecoder.MXP_FLAG already set");
    } else {
      filled.set(36);
    }

    mxpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType1(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_1 already set");
    } else {
      filled.set(37);
    }

    mxpType1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType2(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_2 already set");
    } else {
      filled.set(38);
    }

    mxpType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType3(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_3 already set");
    } else {
      filled.set(39);
    }

    mxpType3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType4(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_4 already set");
    } else {
      filled.set(40);
    }

    mxpType4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType5(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_5 already set");
    } else {
      filled.set(41);
    }

    mxpType5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace nbAdded(final UnsignedByte b) {
    if (filled.get(42)) {
      throw new IllegalStateException("instdecoder.NB_ADDED already set");
    } else {
      filled.set(42);
    }

    nbAdded.put(b.toByte());

    return this;
  }

  public Trace nbRemoved(final UnsignedByte b) {
    if (filled.get(43)) {
      throw new IllegalStateException("instdecoder.NB_REMOVED already set");
    } else {
      filled.set(43);
    }

    nbRemoved.put(b.toByte());

    return this;
  }

  public Trace opcode(final Bytes b) {
    if (filled.get(44)) {
      throw new IllegalStateException("instdecoder.OPCODE already set");
    } else {
      filled.set(44);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      opcode.put((byte) 0);
    }
    opcode.put(b.toArrayUnsafe());

    return this;
  }

  public Trace patternCall(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("instdecoder.PATTERN_CALL already set");
    } else {
      filled.set(45);
    }

    patternCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternCopy(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("instdecoder.PATTERN_COPY already set");
    } else {
      filled.set(46);
    }

    patternCopy.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternCreate(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("instdecoder.PATTERN_CREATE already set");
    } else {
      filled.set(47);
    }

    patternCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternDup(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("instdecoder.PATTERN_DUP already set");
    } else {
      filled.set(48);
    }

    patternDup.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternLoadStore(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("instdecoder.PATTERN_LOAD_STORE already set");
    } else {
      filled.set(49);
    }

    patternLoadStore.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternLog(final Boolean b) {
    if (filled.get(50)) {
      throw new IllegalStateException("instdecoder.PATTERN_LOG already set");
    } else {
      filled.set(50);
    }

    patternLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternOneOne(final Boolean b) {
    if (filled.get(51)) {
      throw new IllegalStateException("instdecoder.PATTERN_ONE_ONE already set");
    } else {
      filled.set(51);
    }

    patternOneOne.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternOneZero(final Boolean b) {
    if (filled.get(52)) {
      throw new IllegalStateException("instdecoder.PATTERN_ONE_ZERO already set");
    } else {
      filled.set(52);
    }

    patternOneZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternSwap(final Boolean b) {
    if (filled.get(53)) {
      throw new IllegalStateException("instdecoder.PATTERN_SWAP already set");
    } else {
      filled.set(53);
    }

    patternSwap.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternThreeOne(final Boolean b) {
    if (filled.get(54)) {
      throw new IllegalStateException("instdecoder.PATTERN_THREE_ONE already set");
    } else {
      filled.set(54);
    }

    patternThreeOne.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternTwoOne(final Boolean b) {
    if (filled.get(55)) {
      throw new IllegalStateException("instdecoder.PATTERN_TWO_ONE already set");
    } else {
      filled.set(55);
    }

    patternTwoOne.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternTwoZero(final Boolean b) {
    if (filled.get(56)) {
      throw new IllegalStateException("instdecoder.PATTERN_TWO_ZERO already set");
    } else {
      filled.set(56);
    }

    patternTwoZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternZeroOne(final Boolean b) {
    if (filled.get(57)) {
      throw new IllegalStateException("instdecoder.PATTERN_ZERO_ONE already set");
    } else {
      filled.set(57);
    }

    patternZeroOne.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace patternZeroZero(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("instdecoder.PATTERN_ZERO_ZERO already set");
    } else {
      filled.set(58);
    }

    patternZeroZero.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramEnabled(final Boolean b) {
    if (filled.get(59)) {
      throw new IllegalStateException("instdecoder.RAM_ENABLED already set");
    } else {
      filled.set(59);
    }

    ramEnabled.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceBlakeData(final Boolean b) {
    if (filled.get(60)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_BLAKE_DATA already set");
    } else {
      filled.set(60);
    }

    ramSourceBlakeData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceEcData(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_EC_DATA already set");
    } else {
      filled.set(61);
    }

    ramSourceEcData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceEcInfo(final Boolean b) {
    if (filled.get(62)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_EC_INFO already set");
    } else {
      filled.set(62);
    }

    ramSourceEcInfo.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceHashData(final Boolean b) {
    if (filled.get(63)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_HASH_DATA already set");
    } else {
      filled.set(63);
    }

    ramSourceHashData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceHashInfo(final Boolean b) {
    if (filled.get(64)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_HASH_INFO already set");
    } else {
      filled.set(64);
    }

    ramSourceHashInfo.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceLogData(final Boolean b) {
    if (filled.get(65)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_LOG_DATA already set");
    } else {
      filled.set(65);
    }

    ramSourceLogData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceModexpData(final Boolean b) {
    if (filled.get(66)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_MODEXP_DATA already set");
    } else {
      filled.set(66);
    }

    ramSourceModexpData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceRam(final Boolean b) {
    if (filled.get(67)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_RAM already set");
    } else {
      filled.set(67);
    }

    ramSourceRam.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceRom(final Boolean b) {
    if (filled.get(68)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_ROM already set");
    } else {
      filled.set(68);
    }

    ramSourceRom.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceStack(final Boolean b) {
    if (filled.get(69)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_STACK already set");
    } else {
      filled.set(69);
    }

    ramSourceStack.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramSourceTxnData(final Boolean b) {
    if (filled.get(70)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_TXN_DATA already set");
    } else {
      filled.set(70);
    }

    ramSourceTxnData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetBlakeData(final Boolean b) {
    if (filled.get(71)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_BLAKE_DATA already set");
    } else {
      filled.set(71);
    }

    ramTargetBlakeData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetEcData(final Boolean b) {
    if (filled.get(72)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_EC_DATA already set");
    } else {
      filled.set(72);
    }

    ramTargetEcData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetEcInfo(final Boolean b) {
    if (filled.get(73)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_EC_INFO already set");
    } else {
      filled.set(73);
    }

    ramTargetEcInfo.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetHashData(final Boolean b) {
    if (filled.get(74)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_HASH_DATA already set");
    } else {
      filled.set(74);
    }

    ramTargetHashData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetHashInfo(final Boolean b) {
    if (filled.get(75)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_HASH_INFO already set");
    } else {
      filled.set(75);
    }

    ramTargetHashInfo.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetLogData(final Boolean b) {
    if (filled.get(76)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_LOG_DATA already set");
    } else {
      filled.set(76);
    }

    ramTargetLogData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetModexpData(final Boolean b) {
    if (filled.get(77)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_MODEXP_DATA already set");
    } else {
      filled.set(77);
    }

    ramTargetModexpData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetRam(final Boolean b) {
    if (filled.get(78)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_RAM already set");
    } else {
      filled.set(78);
    }

    ramTargetRam.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetRom(final Boolean b) {
    if (filled.get(79)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_ROM already set");
    } else {
      filled.set(79);
    }

    ramTargetRom.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetStack(final Boolean b) {
    if (filled.get(80)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_STACK already set");
    } else {
      filled.set(80);
    }

    ramTargetStack.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace ramTargetTxnData(final Boolean b) {
    if (filled.get(81)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_TXN_DATA already set");
    } else {
      filled.set(81);
    }

    ramTargetTxnData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace staticFlag(final Boolean b) {
    if (filled.get(82)) {
      throw new IllegalStateException("instdecoder.STATIC_FLAG already set");
    } else {
      filled.set(82);
    }

    staticFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace staticGas(final Bytes b) {
    if (filled.get(83)) {
      throw new IllegalStateException("instdecoder.STATIC_GAS already set");
    } else {
      filled.set(83);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      staticGas.put((byte) 0);
    }
    staticGas.put(b.toArrayUnsafe());

    return this;
  }

  public Trace twoLineInstruction(final Boolean b) {
    if (filled.get(84)) {
      throw new IllegalStateException("instdecoder.TWO_LINE_INSTRUCTION already set");
    } else {
      filled.set(84);
    }

    twoLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException(
          "instdecoder.ADDRESS_TRIMMING_INSTRUCTION has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("instdecoder.ALPHA has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_BYTE has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_WORD has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("instdecoder.DELTA has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("instdecoder.FAMILY_ACCOUNT has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("instdecoder.FAMILY_ADD has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("instdecoder.FAMILY_BATCH has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("instdecoder.FAMILY_BIN has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("instdecoder.FAMILY_CALL has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("instdecoder.FAMILY_CONTEXT has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("instdecoder.FAMILY_COPY has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("instdecoder.FAMILY_CREATE has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("instdecoder.FAMILY_DUP has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("instdecoder.FAMILY_EXT has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("instdecoder.FAMILY_HALT has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("instdecoder.FAMILY_INVALID has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("instdecoder.FAMILY_JUMP has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("instdecoder.FAMILY_KEC has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("instdecoder.FAMILY_LOG has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("instdecoder.FAMILY_MACHINE_STATE has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("instdecoder.FAMILY_MOD has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("instdecoder.FAMILY_MUL has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("instdecoder.FAMILY_PUSH_POP has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("instdecoder.FAMILY_SHF has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("instdecoder.FAMILY_STACK_RAM has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("instdecoder.FAMILY_STORAGE has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("instdecoder.FAMILY_SWAP has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("instdecoder.FAMILY_TRANSACTION has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("instdecoder.FAMILY_WCP has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("instdecoder.FLAG_1 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("instdecoder.FLAG_2 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("instdecoder.FLAG_3 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("instdecoder.FLAG_4 has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("instdecoder.IS_JUMPDEST has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("instdecoder.IS_PUSH has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("instdecoder.MXP_FLAG has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_1 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_2 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_3 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_4 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_5 has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("instdecoder.NB_ADDED has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("instdecoder.NB_REMOVED has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("instdecoder.OPCODE has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("instdecoder.PATTERN_CALL has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("instdecoder.PATTERN_COPY has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("instdecoder.PATTERN_CREATE has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("instdecoder.PATTERN_DUP has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("instdecoder.PATTERN_LOAD_STORE has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("instdecoder.PATTERN_LOG has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("instdecoder.PATTERN_ONE_ONE has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("instdecoder.PATTERN_ONE_ZERO has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("instdecoder.PATTERN_SWAP has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("instdecoder.PATTERN_THREE_ONE has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("instdecoder.PATTERN_TWO_ONE has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("instdecoder.PATTERN_TWO_ZERO has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("instdecoder.PATTERN_ZERO_ONE has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("instdecoder.PATTERN_ZERO_ZERO has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("instdecoder.RAM_ENABLED has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_BLAKE_DATA has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_EC_DATA has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_EC_INFO has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_HASH_DATA has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_HASH_INFO has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_LOG_DATA has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_MODEXP_DATA has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_RAM has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_ROM has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_STACK has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException("instdecoder.RAM_SOURCE_TXN_DATA has not been filled");
    }

    if (!filled.get(71)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_BLAKE_DATA has not been filled");
    }

    if (!filled.get(72)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_EC_DATA has not been filled");
    }

    if (!filled.get(73)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_EC_INFO has not been filled");
    }

    if (!filled.get(74)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_HASH_DATA has not been filled");
    }

    if (!filled.get(75)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_HASH_INFO has not been filled");
    }

    if (!filled.get(76)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_LOG_DATA has not been filled");
    }

    if (!filled.get(77)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_MODEXP_DATA has not been filled");
    }

    if (!filled.get(78)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_RAM has not been filled");
    }

    if (!filled.get(79)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_ROM has not been filled");
    }

    if (!filled.get(80)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_STACK has not been filled");
    }

    if (!filled.get(81)) {
      throw new IllegalStateException("instdecoder.RAM_TARGET_TXN_DATA has not been filled");
    }

    if (!filled.get(82)) {
      throw new IllegalStateException("instdecoder.STATIC_FLAG has not been filled");
    }

    if (!filled.get(83)) {
      throw new IllegalStateException("instdecoder.STATIC_GAS has not been filled");
    }

    if (!filled.get(84)) {
      throw new IllegalStateException("instdecoder.TWO_LINE_INSTRUCTION has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      addressTrimmingInstruction.position(addressTrimmingInstruction.position() + 1);
    }

    if (!filled.get(1)) {
      alpha.position(alpha.position() + 1);
    }

    if (!filled.get(2)) {
      billingPerByte.position(billingPerByte.position() + 32);
    }

    if (!filled.get(3)) {
      billingPerWord.position(billingPerWord.position() + 32);
    }

    if (!filled.get(4)) {
      delta.position(delta.position() + 1);
    }

    if (!filled.get(5)) {
      familyAccount.position(familyAccount.position() + 1);
    }

    if (!filled.get(6)) {
      familyAdd.position(familyAdd.position() + 1);
    }

    if (!filled.get(7)) {
      familyBatch.position(familyBatch.position() + 1);
    }

    if (!filled.get(8)) {
      familyBin.position(familyBin.position() + 1);
    }

    if (!filled.get(9)) {
      familyCall.position(familyCall.position() + 1);
    }

    if (!filled.get(10)) {
      familyContext.position(familyContext.position() + 1);
    }

    if (!filled.get(11)) {
      familyCopy.position(familyCopy.position() + 1);
    }

    if (!filled.get(12)) {
      familyCreate.position(familyCreate.position() + 1);
    }

    if (!filled.get(13)) {
      familyDup.position(familyDup.position() + 1);
    }

    if (!filled.get(14)) {
      familyExt.position(familyExt.position() + 1);
    }

    if (!filled.get(15)) {
      familyHalt.position(familyHalt.position() + 1);
    }

    if (!filled.get(16)) {
      familyInvalid.position(familyInvalid.position() + 1);
    }

    if (!filled.get(17)) {
      familyJump.position(familyJump.position() + 1);
    }

    if (!filled.get(18)) {
      familyKec.position(familyKec.position() + 1);
    }

    if (!filled.get(19)) {
      familyLog.position(familyLog.position() + 1);
    }

    if (!filled.get(20)) {
      familyMachineState.position(familyMachineState.position() + 1);
    }

    if (!filled.get(21)) {
      familyMod.position(familyMod.position() + 1);
    }

    if (!filled.get(22)) {
      familyMul.position(familyMul.position() + 1);
    }

    if (!filled.get(23)) {
      familyPushPop.position(familyPushPop.position() + 1);
    }

    if (!filled.get(24)) {
      familyShf.position(familyShf.position() + 1);
    }

    if (!filled.get(25)) {
      familyStackRam.position(familyStackRam.position() + 1);
    }

    if (!filled.get(26)) {
      familyStorage.position(familyStorage.position() + 1);
    }

    if (!filled.get(27)) {
      familySwap.position(familySwap.position() + 1);
    }

    if (!filled.get(28)) {
      familyTransaction.position(familyTransaction.position() + 1);
    }

    if (!filled.get(29)) {
      familyWcp.position(familyWcp.position() + 1);
    }

    if (!filled.get(30)) {
      flag1.position(flag1.position() + 1);
    }

    if (!filled.get(31)) {
      flag2.position(flag2.position() + 1);
    }

    if (!filled.get(32)) {
      flag3.position(flag3.position() + 1);
    }

    if (!filled.get(33)) {
      flag4.position(flag4.position() + 1);
    }

    if (!filled.get(34)) {
      isJumpdest.position(isJumpdest.position() + 1);
    }

    if (!filled.get(35)) {
      isPush.position(isPush.position() + 1);
    }

    if (!filled.get(36)) {
      mxpFlag.position(mxpFlag.position() + 1);
    }

    if (!filled.get(37)) {
      mxpType1.position(mxpType1.position() + 1);
    }

    if (!filled.get(38)) {
      mxpType2.position(mxpType2.position() + 1);
    }

    if (!filled.get(39)) {
      mxpType3.position(mxpType3.position() + 1);
    }

    if (!filled.get(40)) {
      mxpType4.position(mxpType4.position() + 1);
    }

    if (!filled.get(41)) {
      mxpType5.position(mxpType5.position() + 1);
    }

    if (!filled.get(42)) {
      nbAdded.position(nbAdded.position() + 1);
    }

    if (!filled.get(43)) {
      nbRemoved.position(nbRemoved.position() + 1);
    }

    if (!filled.get(44)) {
      opcode.position(opcode.position() + 32);
    }

    if (!filled.get(45)) {
      patternCall.position(patternCall.position() + 1);
    }

    if (!filled.get(46)) {
      patternCopy.position(patternCopy.position() + 1);
    }

    if (!filled.get(47)) {
      patternCreate.position(patternCreate.position() + 1);
    }

    if (!filled.get(48)) {
      patternDup.position(patternDup.position() + 1);
    }

    if (!filled.get(49)) {
      patternLoadStore.position(patternLoadStore.position() + 1);
    }

    if (!filled.get(50)) {
      patternLog.position(patternLog.position() + 1);
    }

    if (!filled.get(51)) {
      patternOneOne.position(patternOneOne.position() + 1);
    }

    if (!filled.get(52)) {
      patternOneZero.position(patternOneZero.position() + 1);
    }

    if (!filled.get(53)) {
      patternSwap.position(patternSwap.position() + 1);
    }

    if (!filled.get(54)) {
      patternThreeOne.position(patternThreeOne.position() + 1);
    }

    if (!filled.get(55)) {
      patternTwoOne.position(patternTwoOne.position() + 1);
    }

    if (!filled.get(56)) {
      patternTwoZero.position(patternTwoZero.position() + 1);
    }

    if (!filled.get(57)) {
      patternZeroOne.position(patternZeroOne.position() + 1);
    }

    if (!filled.get(58)) {
      patternZeroZero.position(patternZeroZero.position() + 1);
    }

    if (!filled.get(59)) {
      ramEnabled.position(ramEnabled.position() + 1);
    }

    if (!filled.get(60)) {
      ramSourceBlakeData.position(ramSourceBlakeData.position() + 1);
    }

    if (!filled.get(61)) {
      ramSourceEcData.position(ramSourceEcData.position() + 1);
    }

    if (!filled.get(62)) {
      ramSourceEcInfo.position(ramSourceEcInfo.position() + 1);
    }

    if (!filled.get(63)) {
      ramSourceHashData.position(ramSourceHashData.position() + 1);
    }

    if (!filled.get(64)) {
      ramSourceHashInfo.position(ramSourceHashInfo.position() + 1);
    }

    if (!filled.get(65)) {
      ramSourceLogData.position(ramSourceLogData.position() + 1);
    }

    if (!filled.get(66)) {
      ramSourceModexpData.position(ramSourceModexpData.position() + 1);
    }

    if (!filled.get(67)) {
      ramSourceRam.position(ramSourceRam.position() + 1);
    }

    if (!filled.get(68)) {
      ramSourceRom.position(ramSourceRom.position() + 1);
    }

    if (!filled.get(69)) {
      ramSourceStack.position(ramSourceStack.position() + 1);
    }

    if (!filled.get(70)) {
      ramSourceTxnData.position(ramSourceTxnData.position() + 1);
    }

    if (!filled.get(71)) {
      ramTargetBlakeData.position(ramTargetBlakeData.position() + 1);
    }

    if (!filled.get(72)) {
      ramTargetEcData.position(ramTargetEcData.position() + 1);
    }

    if (!filled.get(73)) {
      ramTargetEcInfo.position(ramTargetEcInfo.position() + 1);
    }

    if (!filled.get(74)) {
      ramTargetHashData.position(ramTargetHashData.position() + 1);
    }

    if (!filled.get(75)) {
      ramTargetHashInfo.position(ramTargetHashInfo.position() + 1);
    }

    if (!filled.get(76)) {
      ramTargetLogData.position(ramTargetLogData.position() + 1);
    }

    if (!filled.get(77)) {
      ramTargetModexpData.position(ramTargetModexpData.position() + 1);
    }

    if (!filled.get(78)) {
      ramTargetRam.position(ramTargetRam.position() + 1);
    }

    if (!filled.get(79)) {
      ramTargetRom.position(ramTargetRom.position() + 1);
    }

    if (!filled.get(80)) {
      ramTargetStack.position(ramTargetStack.position() + 1);
    }

    if (!filled.get(81)) {
      ramTargetTxnData.position(ramTargetTxnData.position() + 1);
    }

    if (!filled.get(82)) {
      staticFlag.position(staticFlag.position() + 1);
    }

    if (!filled.get(83)) {
      staticGas.position(staticGas.position() + 32);
    }

    if (!filled.get(84)) {
      twoLineInstruction.position(twoLineInstruction.position() + 1);
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public void build() {
    if (!filled.isEmpty()) {
      throw new IllegalStateException("Cannot build trace with a non-validated row.");
    }
  }
}
