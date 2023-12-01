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

package net.consensys.linea.zktracer.module.rom;

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

  private final MappedByteBuffer acc;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexInfty;
  private final MappedByteBuffer codeSize;
  private final MappedByteBuffer codesizeReached;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer counterMax;
  private final MappedByteBuffer counterPush;
  private final MappedByteBuffer index;
  private final MappedByteBuffer isPush;
  private final MappedByteBuffer isPushData;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nBytesAcc;
  private final MappedByteBuffer opcode;
  private final MappedByteBuffer paddedBytecodeByte;
  private final MappedByteBuffer programmeCounter;
  private final MappedByteBuffer pushFunnelBit;
  private final MappedByteBuffer pushParameter;
  private final MappedByteBuffer pushValueAcc;
  private final MappedByteBuffer pushValueHigh;
  private final MappedByteBuffer pushValueLow;
  private final MappedByteBuffer validJumpDestination;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("rom.ACC", 32, length),
        new ColumnHeader("rom.CODE_FRAGMENT_INDEX", 32, length),
        new ColumnHeader("rom.CODE_FRAGMENT_INDEX_INFTY", 32, length),
        new ColumnHeader("rom.CODE_SIZE", 32, length),
        new ColumnHeader("rom.CODESIZE_REACHED", 1, length),
        new ColumnHeader("rom.COUNTER", 32, length),
        new ColumnHeader("rom.COUNTER_MAX", 32, length),
        new ColumnHeader("rom.COUNTER_PUSH", 32, length),
        new ColumnHeader("rom.INDEX", 32, length),
        new ColumnHeader("rom.IS_PUSH", 1, length),
        new ColumnHeader("rom.IS_PUSH_DATA", 1, length),
        new ColumnHeader("rom.LIMB", 32, length),
        new ColumnHeader("rom.nBYTES", 32, length),
        new ColumnHeader("rom.nBYTES_ACC", 32, length),
        new ColumnHeader("rom.OPCODE", 1, length),
        new ColumnHeader("rom.PADDED_BYTECODE_BYTE", 1, length),
        new ColumnHeader("rom.PROGRAMME_COUNTER", 32, length),
        new ColumnHeader("rom.PUSH_FUNNEL_BIT", 1, length),
        new ColumnHeader("rom.PUSH_PARAMETER", 32, length),
        new ColumnHeader("rom.PUSH_VALUE_ACC", 32, length),
        new ColumnHeader("rom.PUSH_VALUE_HIGH", 32, length),
        new ColumnHeader("rom.PUSH_VALUE_LOW", 32, length),
        new ColumnHeader("rom.VALID_JUMP_DESTINATION", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc = buffers.get(0);
    this.codeFragmentIndex = buffers.get(1);
    this.codeFragmentIndexInfty = buffers.get(2);
    this.codeSize = buffers.get(3);
    this.codesizeReached = buffers.get(4);
    this.counter = buffers.get(5);
    this.counterMax = buffers.get(6);
    this.counterPush = buffers.get(7);
    this.index = buffers.get(8);
    this.isPush = buffers.get(9);
    this.isPushData = buffers.get(10);
    this.limb = buffers.get(11);
    this.nBytes = buffers.get(12);
    this.nBytesAcc = buffers.get(13);
    this.opcode = buffers.get(14);
    this.paddedBytecodeByte = buffers.get(15);
    this.programmeCounter = buffers.get(16);
    this.pushFunnelBit = buffers.get(17);
    this.pushParameter = buffers.get(18);
    this.pushValueAcc = buffers.get(19);
    this.pushValueHigh = buffers.get(20);
    this.pushValueLow = buffers.get(21);
    this.validJumpDestination = buffers.get(22);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rom.ACC already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      acc.put((byte) 0);
    }
    acc.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndex(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndex.put((byte) 0);
    }
    codeFragmentIndex.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndexInfty(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX_INFTY already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeFragmentIndexInfty.put((byte) 0);
    }
    codeFragmentIndexInfty.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeSize(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rom.CODE_SIZE already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      codeSize.put((byte) 0);
    }
    codeSize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codesizeReached(final Boolean b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rom.CODESIZE_REACHED already set");
    } else {
      filled.set(1);
    }

    codesizeReached.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace counter(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rom.COUNTER already set");
    } else {
      filled.set(5);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counter.put((byte) 0);
    }
    counter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace counterMax(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rom.COUNTER_MAX already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counterMax.put((byte) 0);
    }
    counterMax.put(b.toArrayUnsafe());

    return this;
  }

  public Trace counterPush(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rom.COUNTER_PUSH already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counterPush.put((byte) 0);
    }
    counterPush.put(b.toArrayUnsafe());

    return this;
  }

  public Trace index(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rom.INDEX already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      index.put((byte) 0);
    }
    index.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isPush(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rom.IS_PUSH already set");
    } else {
      filled.set(9);
    }

    isPush.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPushData(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rom.IS_PUSH_DATA already set");
    } else {
      filled.set(10);
    }

    isPushData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rom.LIMB already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nBytes(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rom.nBYTES already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nBytes.put((byte) 0);
    }
    nBytes.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nBytesAcc(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rom.nBYTES_ACC already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nBytesAcc.put((byte) 0);
    }
    nBytesAcc.put(b.toArrayUnsafe());

    return this;
  }

  public Trace opcode(final UnsignedByte b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rom.OPCODE already set");
    } else {
      filled.set(12);
    }

    opcode.put(b.toByte());

    return this;
  }

  public Trace paddedBytecodeByte(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rom.PADDED_BYTECODE_BYTE already set");
    } else {
      filled.set(13);
    }

    paddedBytecodeByte.put(b.toByte());

    return this;
  }

  public Trace programmeCounter(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rom.PROGRAMME_COUNTER already set");
    } else {
      filled.set(14);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      programmeCounter.put((byte) 0);
    }
    programmeCounter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pushFunnelBit(final Boolean b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rom.PUSH_FUNNEL_BIT already set");
    } else {
      filled.set(15);
    }

    pushFunnelBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pushParameter(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rom.PUSH_PARAMETER already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushParameter.put((byte) 0);
    }
    pushParameter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pushValueAcc(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rom.PUSH_VALUE_ACC already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushValueAcc.put((byte) 0);
    }
    pushValueAcc.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pushValueHigh(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rom.PUSH_VALUE_HIGH already set");
    } else {
      filled.set(18);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushValueHigh.put((byte) 0);
    }
    pushValueHigh.put(b.toArrayUnsafe());

    return this;
  }

  public Trace pushValueLow(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rom.PUSH_VALUE_LOW already set");
    } else {
      filled.set(19);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      pushValueLow.put((byte) 0);
    }
    pushValueLow.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validJumpDestination(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rom.VALID_JUMP_DESTINATION already set");
    } else {
      filled.set(20);
    }

    validJumpDestination.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rom.ACC has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX_INFTY has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rom.CODE_SIZE has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rom.CODESIZE_REACHED has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rom.COUNTER has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rom.COUNTER_MAX has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rom.COUNTER_PUSH has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rom.INDEX has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rom.IS_PUSH has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rom.IS_PUSH_DATA has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rom.LIMB has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rom.nBYTES has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rom.nBYTES_ACC has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rom.OPCODE has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rom.PADDED_BYTECODE_BYTE has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rom.PROGRAMME_COUNTER has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rom.PUSH_FUNNEL_BIT has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rom.PUSH_PARAMETER has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rom.PUSH_VALUE_ACC has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rom.PUSH_VALUE_HIGH has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rom.PUSH_VALUE_LOW has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rom.VALID_JUMP_DESTINATION has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc.position(acc.position() + 32);
    }

    if (!filled.get(2)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 32);
    }

    if (!filled.get(3)) {
      codeFragmentIndexInfty.position(codeFragmentIndexInfty.position() + 32);
    }

    if (!filled.get(4)) {
      codeSize.position(codeSize.position() + 32);
    }

    if (!filled.get(1)) {
      codesizeReached.position(codesizeReached.position() + 1);
    }

    if (!filled.get(5)) {
      counter.position(counter.position() + 32);
    }

    if (!filled.get(6)) {
      counterMax.position(counterMax.position() + 32);
    }

    if (!filled.get(7)) {
      counterPush.position(counterPush.position() + 32);
    }

    if (!filled.get(8)) {
      index.position(index.position() + 32);
    }

    if (!filled.get(9)) {
      isPush.position(isPush.position() + 1);
    }

    if (!filled.get(10)) {
      isPushData.position(isPushData.position() + 1);
    }

    if (!filled.get(11)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(21)) {
      nBytes.position(nBytes.position() + 32);
    }

    if (!filled.get(22)) {
      nBytesAcc.position(nBytesAcc.position() + 32);
    }

    if (!filled.get(12)) {
      opcode.position(opcode.position() + 1);
    }

    if (!filled.get(13)) {
      paddedBytecodeByte.position(paddedBytecodeByte.position() + 1);
    }

    if (!filled.get(14)) {
      programmeCounter.position(programmeCounter.position() + 32);
    }

    if (!filled.get(15)) {
      pushFunnelBit.position(pushFunnelBit.position() + 1);
    }

    if (!filled.get(16)) {
      pushParameter.position(pushParameter.position() + 32);
    }

    if (!filled.get(17)) {
      pushValueAcc.position(pushValueAcc.position() + 32);
    }

    if (!filled.get(18)) {
      pushValueHigh.position(pushValueHigh.position() + 32);
    }

    if (!filled.get(19)) {
      pushValueLow.position(pushValueLow.position() + 32);
    }

    if (!filled.get(20)) {
      validJumpDestination.position(validJumpDestination.position() + 1);
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
