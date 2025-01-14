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

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
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
  private final MappedByteBuffer opcode;
  private final MappedByteBuffer staticFlag;
  private final MappedByteBuffer staticGas;
  private final MappedByteBuffer twoLineInstruction;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("instdecoder.ALPHA", 1, length));
      headers.add(new ColumnHeader("instdecoder.BILLING_PER_BYTE", 1, length));
      headers.add(new ColumnHeader("instdecoder.BILLING_PER_WORD", 1, length));
      headers.add(new ColumnHeader("instdecoder.DELTA", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_ACCOUNT", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_ADD", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_BATCH", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_BIN", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_CALL", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_CONTEXT", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_COPY", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_CREATE", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_DUP", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_EXT", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_HALT", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_INVALID", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_JUMP", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_KEC", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_LOG", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_MACHINE_STATE", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_MOD", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_MUL", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_PUSH_POP", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_SHF", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_STACK_RAM", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_STORAGE", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_SWAP", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_TRANSACTION", 1, length));
      headers.add(new ColumnHeader("instdecoder.FAMILY_WCP", 1, length));
      headers.add(new ColumnHeader("instdecoder.FLAG_1", 1, length));
      headers.add(new ColumnHeader("instdecoder.FLAG_2", 1, length));
      headers.add(new ColumnHeader("instdecoder.FLAG_3", 1, length));
      headers.add(new ColumnHeader("instdecoder.FLAG_4", 1, length));
      headers.add(new ColumnHeader("instdecoder.IS_JUMPDEST", 1, length));
      headers.add(new ColumnHeader("instdecoder.IS_PUSH", 1, length));
      headers.add(new ColumnHeader("instdecoder.MXP_FLAG", 1, length));
      headers.add(new ColumnHeader("instdecoder.MXP_TYPE_1", 1, length));
      headers.add(new ColumnHeader("instdecoder.MXP_TYPE_2", 1, length));
      headers.add(new ColumnHeader("instdecoder.MXP_TYPE_3", 1, length));
      headers.add(new ColumnHeader("instdecoder.MXP_TYPE_4", 1, length));
      headers.add(new ColumnHeader("instdecoder.MXP_TYPE_5", 1, length));
      headers.add(new ColumnHeader("instdecoder.OPCODE", 32, length));
      headers.add(new ColumnHeader("instdecoder.STATIC_FLAG", 1, length));
      headers.add(new ColumnHeader("instdecoder.STATIC_GAS", 4, length));
      headers.add(new ColumnHeader("instdecoder.TWO_LINE_INSTRUCTION", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.alpha = buffers.get(0);
    this.billingPerByte = buffers.get(1);
    this.billingPerWord = buffers.get(2);
    this.delta = buffers.get(3);
    this.familyAccount = buffers.get(4);
    this.familyAdd = buffers.get(5);
    this.familyBatch = buffers.get(6);
    this.familyBin = buffers.get(7);
    this.familyCall = buffers.get(8);
    this.familyContext = buffers.get(9);
    this.familyCopy = buffers.get(10);
    this.familyCreate = buffers.get(11);
    this.familyDup = buffers.get(12);
    this.familyExt = buffers.get(13);
    this.familyHalt = buffers.get(14);
    this.familyInvalid = buffers.get(15);
    this.familyJump = buffers.get(16);
    this.familyKec = buffers.get(17);
    this.familyLog = buffers.get(18);
    this.familyMachineState = buffers.get(19);
    this.familyMod = buffers.get(20);
    this.familyMul = buffers.get(21);
    this.familyPushPop = buffers.get(22);
    this.familyShf = buffers.get(23);
    this.familyStackRam = buffers.get(24);
    this.familyStorage = buffers.get(25);
    this.familySwap = buffers.get(26);
    this.familyTransaction = buffers.get(27);
    this.familyWcp = buffers.get(28);
    this.flag1 = buffers.get(29);
    this.flag2 = buffers.get(30);
    this.flag3 = buffers.get(31);
    this.flag4 = buffers.get(32);
    this.isJumpdest = buffers.get(33);
    this.isPush = buffers.get(34);
    this.mxpFlag = buffers.get(35);
    this.mxpType1 = buffers.get(36);
    this.mxpType2 = buffers.get(37);
    this.mxpType3 = buffers.get(38);
    this.mxpType4 = buffers.get(39);
    this.mxpType5 = buffers.get(40);
    this.opcode = buffers.get(41);
    this.staticFlag = buffers.get(42);
    this.staticGas = buffers.get(43);
    this.twoLineInstruction = buffers.get(44);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace alpha(final UnsignedByte b) {
    if (filled.get(0)) {
      throw new IllegalStateException("instdecoder.ALPHA already set");
    } else {
      filled.set(0);
    }

    alpha.put(b.toByte());

    return this;
  }

  public Trace billingPerByte(final UnsignedByte b) {
    if (filled.get(1)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_BYTE already set");
    } else {
      filled.set(1);
    }

    billingPerByte.put(b.toByte());

    return this;
  }

  public Trace billingPerWord(final UnsignedByte b) {
    if (filled.get(2)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_WORD already set");
    } else {
      filled.set(2);
    }

    billingPerWord.put(b.toByte());

    return this;
  }

  public Trace delta(final UnsignedByte b) {
    if (filled.get(3)) {
      throw new IllegalStateException("instdecoder.DELTA already set");
    } else {
      filled.set(3);
    }

    delta.put(b.toByte());

    return this;
  }

  public Trace familyAccount(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("instdecoder.FAMILY_ACCOUNT already set");
    } else {
      filled.set(4);
    }

    familyAccount.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyAdd(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("instdecoder.FAMILY_ADD already set");
    } else {
      filled.set(5);
    }

    familyAdd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyBatch(final Boolean b) {
    if (filled.get(6)) {
      throw new IllegalStateException("instdecoder.FAMILY_BATCH already set");
    } else {
      filled.set(6);
    }

    familyBatch.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyBin(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("instdecoder.FAMILY_BIN already set");
    } else {
      filled.set(7);
    }

    familyBin.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyCall(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("instdecoder.FAMILY_CALL already set");
    } else {
      filled.set(8);
    }

    familyCall.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyContext(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("instdecoder.FAMILY_CONTEXT already set");
    } else {
      filled.set(9);
    }

    familyContext.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyCopy(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("instdecoder.FAMILY_COPY already set");
    } else {
      filled.set(10);
    }

    familyCopy.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyCreate(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("instdecoder.FAMILY_CREATE already set");
    } else {
      filled.set(11);
    }

    familyCreate.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyDup(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("instdecoder.FAMILY_DUP already set");
    } else {
      filled.set(12);
    }

    familyDup.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyExt(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("instdecoder.FAMILY_EXT already set");
    } else {
      filled.set(13);
    }

    familyExt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyHalt(final Boolean b) {
    if (filled.get(14)) {
      throw new IllegalStateException("instdecoder.FAMILY_HALT already set");
    } else {
      filled.set(14);
    }

    familyHalt.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyInvalid(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("instdecoder.FAMILY_INVALID already set");
    } else {
      filled.set(15);
    }

    familyInvalid.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyJump(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("instdecoder.FAMILY_JUMP already set");
    } else {
      filled.set(16);
    }

    familyJump.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyKec(final Boolean b) {
    if (filled.get(17)) {
      throw new IllegalStateException("instdecoder.FAMILY_KEC already set");
    } else {
      filled.set(17);
    }

    familyKec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyLog(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("instdecoder.FAMILY_LOG already set");
    } else {
      filled.set(18);
    }

    familyLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyMachineState(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("instdecoder.FAMILY_MACHINE_STATE already set");
    } else {
      filled.set(19);
    }

    familyMachineState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyMod(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("instdecoder.FAMILY_MOD already set");
    } else {
      filled.set(20);
    }

    familyMod.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyMul(final Boolean b) {
    if (filled.get(21)) {
      throw new IllegalStateException("instdecoder.FAMILY_MUL already set");
    } else {
      filled.set(21);
    }

    familyMul.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyPushPop(final Boolean b) {
    if (filled.get(22)) {
      throw new IllegalStateException("instdecoder.FAMILY_PUSH_POP already set");
    } else {
      filled.set(22);
    }

    familyPushPop.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyShf(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("instdecoder.FAMILY_SHF already set");
    } else {
      filled.set(23);
    }

    familyShf.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyStackRam(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("instdecoder.FAMILY_STACK_RAM already set");
    } else {
      filled.set(24);
    }

    familyStackRam.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyStorage(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("instdecoder.FAMILY_STORAGE already set");
    } else {
      filled.set(25);
    }

    familyStorage.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familySwap(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("instdecoder.FAMILY_SWAP already set");
    } else {
      filled.set(26);
    }

    familySwap.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyTransaction(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("instdecoder.FAMILY_TRANSACTION already set");
    } else {
      filled.set(27);
    }

    familyTransaction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace familyWcp(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("instdecoder.FAMILY_WCP already set");
    } else {
      filled.set(28);
    }

    familyWcp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag1(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("instdecoder.FLAG_1 already set");
    } else {
      filled.set(29);
    }

    flag1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag2(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("instdecoder.FLAG_2 already set");
    } else {
      filled.set(30);
    }

    flag2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag3(final Boolean b) {
    if (filled.get(31)) {
      throw new IllegalStateException("instdecoder.FLAG_3 already set");
    } else {
      filled.set(31);
    }

    flag3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace flag4(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("instdecoder.FLAG_4 already set");
    } else {
      filled.set(32);
    }

    flag4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isJumpdest(final Boolean b) {
    if (filled.get(33)) {
      throw new IllegalStateException("instdecoder.IS_JUMPDEST already set");
    } else {
      filled.set(33);
    }

    isJumpdest.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPush(final Boolean b) {
    if (filled.get(34)) {
      throw new IllegalStateException("instdecoder.IS_PUSH already set");
    } else {
      filled.set(34);
    }

    isPush.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpFlag(final Boolean b) {
    if (filled.get(35)) {
      throw new IllegalStateException("instdecoder.MXP_FLAG already set");
    } else {
      filled.set(35);
    }

    mxpFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType1(final Boolean b) {
    if (filled.get(36)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_1 already set");
    } else {
      filled.set(36);
    }

    mxpType1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType2(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_2 already set");
    } else {
      filled.set(37);
    }

    mxpType2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType3(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_3 already set");
    } else {
      filled.set(38);
    }

    mxpType3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType4(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_4 already set");
    } else {
      filled.set(39);
    }

    mxpType4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace mxpType5(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_5 already set");
    } else {
      filled.set(40);
    }

    mxpType5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace opcode(final Bytes b) {
    if (filled.get(41)) {
      throw new IllegalStateException("instdecoder.OPCODE already set");
    } else {
      filled.set(41);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 256) { throw new IllegalArgumentException("instdecoder.OPCODE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<32; i++) { opcode.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { opcode.put(bs.get(j)); }

    return this;
  }

  public Trace staticFlag(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("instdecoder.STATIC_FLAG already set");
    } else {
      filled.set(42);
    }

    staticFlag.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace staticGas(final long b) {
    if (filled.get(43)) {
      throw new IllegalStateException("instdecoder.STATIC_GAS already set");
    } else {
      filled.set(43);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("instdecoder.STATIC_GAS has invalid value (" + b + ")"); }
    staticGas.put((byte) (b >> 24));
    staticGas.put((byte) (b >> 16));
    staticGas.put((byte) (b >> 8));
    staticGas.put((byte) b);


    return this;
  }

  public Trace twoLineInstruction(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("instdecoder.TWO_LINE_INSTRUCTION already set");
    } else {
      filled.set(44);
    }

    twoLineInstruction.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("instdecoder.ALPHA has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_BYTE has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("instdecoder.BILLING_PER_WORD has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("instdecoder.DELTA has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("instdecoder.FAMILY_ACCOUNT has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("instdecoder.FAMILY_ADD has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("instdecoder.FAMILY_BATCH has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("instdecoder.FAMILY_BIN has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("instdecoder.FAMILY_CALL has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("instdecoder.FAMILY_CONTEXT has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("instdecoder.FAMILY_COPY has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("instdecoder.FAMILY_CREATE has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("instdecoder.FAMILY_DUP has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("instdecoder.FAMILY_EXT has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("instdecoder.FAMILY_HALT has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("instdecoder.FAMILY_INVALID has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("instdecoder.FAMILY_JUMP has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("instdecoder.FAMILY_KEC has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("instdecoder.FAMILY_LOG has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("instdecoder.FAMILY_MACHINE_STATE has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("instdecoder.FAMILY_MOD has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("instdecoder.FAMILY_MUL has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("instdecoder.FAMILY_PUSH_POP has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("instdecoder.FAMILY_SHF has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("instdecoder.FAMILY_STACK_RAM has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("instdecoder.FAMILY_STORAGE has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("instdecoder.FAMILY_SWAP has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("instdecoder.FAMILY_TRANSACTION has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("instdecoder.FAMILY_WCP has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("instdecoder.FLAG_1 has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("instdecoder.FLAG_2 has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("instdecoder.FLAG_3 has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("instdecoder.FLAG_4 has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("instdecoder.IS_JUMPDEST has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("instdecoder.IS_PUSH has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("instdecoder.MXP_FLAG has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_1 has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_2 has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_3 has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_4 has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("instdecoder.MXP_TYPE_5 has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("instdecoder.OPCODE has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("instdecoder.STATIC_FLAG has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("instdecoder.STATIC_GAS has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("instdecoder.TWO_LINE_INSTRUCTION has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      alpha.position(alpha.position() + 1);
    }

    if (!filled.get(1)) {
      billingPerByte.position(billingPerByte.position() + 1);
    }

    if (!filled.get(2)) {
      billingPerWord.position(billingPerWord.position() + 1);
    }

    if (!filled.get(3)) {
      delta.position(delta.position() + 1);
    }

    if (!filled.get(4)) {
      familyAccount.position(familyAccount.position() + 1);
    }

    if (!filled.get(5)) {
      familyAdd.position(familyAdd.position() + 1);
    }

    if (!filled.get(6)) {
      familyBatch.position(familyBatch.position() + 1);
    }

    if (!filled.get(7)) {
      familyBin.position(familyBin.position() + 1);
    }

    if (!filled.get(8)) {
      familyCall.position(familyCall.position() + 1);
    }

    if (!filled.get(9)) {
      familyContext.position(familyContext.position() + 1);
    }

    if (!filled.get(10)) {
      familyCopy.position(familyCopy.position() + 1);
    }

    if (!filled.get(11)) {
      familyCreate.position(familyCreate.position() + 1);
    }

    if (!filled.get(12)) {
      familyDup.position(familyDup.position() + 1);
    }

    if (!filled.get(13)) {
      familyExt.position(familyExt.position() + 1);
    }

    if (!filled.get(14)) {
      familyHalt.position(familyHalt.position() + 1);
    }

    if (!filled.get(15)) {
      familyInvalid.position(familyInvalid.position() + 1);
    }

    if (!filled.get(16)) {
      familyJump.position(familyJump.position() + 1);
    }

    if (!filled.get(17)) {
      familyKec.position(familyKec.position() + 1);
    }

    if (!filled.get(18)) {
      familyLog.position(familyLog.position() + 1);
    }

    if (!filled.get(19)) {
      familyMachineState.position(familyMachineState.position() + 1);
    }

    if (!filled.get(20)) {
      familyMod.position(familyMod.position() + 1);
    }

    if (!filled.get(21)) {
      familyMul.position(familyMul.position() + 1);
    }

    if (!filled.get(22)) {
      familyPushPop.position(familyPushPop.position() + 1);
    }

    if (!filled.get(23)) {
      familyShf.position(familyShf.position() + 1);
    }

    if (!filled.get(24)) {
      familyStackRam.position(familyStackRam.position() + 1);
    }

    if (!filled.get(25)) {
      familyStorage.position(familyStorage.position() + 1);
    }

    if (!filled.get(26)) {
      familySwap.position(familySwap.position() + 1);
    }

    if (!filled.get(27)) {
      familyTransaction.position(familyTransaction.position() + 1);
    }

    if (!filled.get(28)) {
      familyWcp.position(familyWcp.position() + 1);
    }

    if (!filled.get(29)) {
      flag1.position(flag1.position() + 1);
    }

    if (!filled.get(30)) {
      flag2.position(flag2.position() + 1);
    }

    if (!filled.get(31)) {
      flag3.position(flag3.position() + 1);
    }

    if (!filled.get(32)) {
      flag4.position(flag4.position() + 1);
    }

    if (!filled.get(33)) {
      isJumpdest.position(isJumpdest.position() + 1);
    }

    if (!filled.get(34)) {
      isPush.position(isPush.position() + 1);
    }

    if (!filled.get(35)) {
      mxpFlag.position(mxpFlag.position() + 1);
    }

    if (!filled.get(36)) {
      mxpType1.position(mxpType1.position() + 1);
    }

    if (!filled.get(37)) {
      mxpType2.position(mxpType2.position() + 1);
    }

    if (!filled.get(38)) {
      mxpType3.position(mxpType3.position() + 1);
    }

    if (!filled.get(39)) {
      mxpType4.position(mxpType4.position() + 1);
    }

    if (!filled.get(40)) {
      mxpType5.position(mxpType5.position() + 1);
    }

    if (!filled.get(41)) {
      opcode.position(opcode.position() + 32);
    }

    if (!filled.get(42)) {
      staticFlag.position(staticFlag.position() + 1);
    }

    if (!filled.get(43)) {
      staticGas.position(staticGas.position() + 4);
    }

    if (!filled.get(44)) {
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
