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

  private final MappedByteBuffer acc;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexInfty;
  private final MappedByteBuffer codeSize;
  private final MappedByteBuffer codesizeReached;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer counterMax;
  private final MappedByteBuffer counterPush;
  private final MappedByteBuffer index;
  private final MappedByteBuffer isJumpdest;
  private final MappedByteBuffer isPush;
  private final MappedByteBuffer isPushData;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nBytesAcc;
  private final MappedByteBuffer opcode;
  private final MappedByteBuffer paddedBytecodeByte;
  private final MappedByteBuffer programCounter;
  private final MappedByteBuffer pushFunnelBit;
  private final MappedByteBuffer pushParameter;
  private final MappedByteBuffer pushValueAcc;
  private final MappedByteBuffer pushValueHi;
  private final MappedByteBuffer pushValueLo;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("rom.ACC", 16, length));
      headers.add(new ColumnHeader("rom.CODE_FRAGMENT_INDEX", 4, length));
      headers.add(new ColumnHeader("rom.CODE_FRAGMENT_INDEX_INFTY", 4, length));
      headers.add(new ColumnHeader("rom.CODE_SIZE", 4, length));
      headers.add(new ColumnHeader("rom.CODESIZE_REACHED", 1, length));
      headers.add(new ColumnHeader("rom.COUNTER", 1, length));
      headers.add(new ColumnHeader("rom.COUNTER_MAX", 1, length));
      headers.add(new ColumnHeader("rom.COUNTER_PUSH", 1, length));
      headers.add(new ColumnHeader("rom.INDEX", 4, length));
      headers.add(new ColumnHeader("rom.IS_JUMPDEST", 1, length));
      headers.add(new ColumnHeader("rom.IS_PUSH", 1, length));
      headers.add(new ColumnHeader("rom.IS_PUSH_DATA", 1, length));
      headers.add(new ColumnHeader("rom.LIMB", 16, length));
      headers.add(new ColumnHeader("rom.nBYTES", 1, length));
      headers.add(new ColumnHeader("rom.nBYTES_ACC", 1, length));
      headers.add(new ColumnHeader("rom.OPCODE", 1, length));
      headers.add(new ColumnHeader("rom.PADDED_BYTECODE_BYTE", 1, length));
      headers.add(new ColumnHeader("rom.PROGRAM_COUNTER", 4, length));
      headers.add(new ColumnHeader("rom.PUSH_FUNNEL_BIT", 1, length));
      headers.add(new ColumnHeader("rom.PUSH_PARAMETER", 1, length));
      headers.add(new ColumnHeader("rom.PUSH_VALUE_ACC", 16, length));
      headers.add(new ColumnHeader("rom.PUSH_VALUE_HI", 16, length));
      headers.add(new ColumnHeader("rom.PUSH_VALUE_LO", 16, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.acc = buffers.get(0);
    this.codeFragmentIndex = buffers.get(1);
    this.codeFragmentIndexInfty = buffers.get(2);
    this.codeSize = buffers.get(3);
    this.codesizeReached = buffers.get(4);
    this.counter = buffers.get(5);
    this.counterMax = buffers.get(6);
    this.counterPush = buffers.get(7);
    this.index = buffers.get(8);
    this.isJumpdest = buffers.get(9);
    this.isPush = buffers.get(10);
    this.isPushData = buffers.get(11);
    this.limb = buffers.get(12);
    this.nBytes = buffers.get(13);
    this.nBytesAcc = buffers.get(14);
    this.opcode = buffers.get(15);
    this.paddedBytecodeByte = buffers.get(16);
    this.programCounter = buffers.get(17);
    this.pushFunnelBit = buffers.get(18);
    this.pushParameter = buffers.get(19);
    this.pushValueAcc = buffers.get(20);
    this.pushValueHi = buffers.get(21);
    this.pushValueLo = buffers.get(22);
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

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rom.ACC has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc.put(bs.get(j)); }

    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(2);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rom.CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeFragmentIndex.put((byte) (b >> 24));
    codeFragmentIndex.put((byte) (b >> 16));
    codeFragmentIndex.put((byte) (b >> 8));
    codeFragmentIndex.put((byte) b);


    return this;
  }

  public Trace codeFragmentIndexInfty(final long b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rom.CODE_FRAGMENT_INDEX_INFTY already set");
    } else {
      filled.set(3);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rom.CODE_FRAGMENT_INDEX_INFTY has invalid value (" + b + ")"); }
    codeFragmentIndexInfty.put((byte) (b >> 24));
    codeFragmentIndexInfty.put((byte) (b >> 16));
    codeFragmentIndexInfty.put((byte) (b >> 8));
    codeFragmentIndexInfty.put((byte) b);


    return this;
  }

  public Trace codeSize(final long b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rom.CODE_SIZE already set");
    } else {
      filled.set(4);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rom.CODE_SIZE has invalid value (" + b + ")"); }
    codeSize.put((byte) (b >> 24));
    codeSize.put((byte) (b >> 16));
    codeSize.put((byte) (b >> 8));
    codeSize.put((byte) b);


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

  public Trace counter(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rom.COUNTER already set");
    } else {
      filled.set(5);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace counterMax(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rom.COUNTER_MAX already set");
    } else {
      filled.set(6);
    }

    counterMax.put(b.toByte());

    return this;
  }

  public Trace counterPush(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rom.COUNTER_PUSH already set");
    } else {
      filled.set(7);
    }

    counterPush.put(b.toByte());

    return this;
  }

  public Trace index(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rom.INDEX already set");
    } else {
      filled.set(8);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rom.INDEX has invalid value (" + b + ")"); }
    index.put((byte) (b >> 24));
    index.put((byte) (b >> 16));
    index.put((byte) (b >> 8));
    index.put((byte) b);


    return this;
  }

  public Trace isJumpdest(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rom.IS_JUMPDEST already set");
    } else {
      filled.set(9);
    }

    isJumpdest.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPush(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rom.IS_PUSH already set");
    } else {
      filled.set(10);
    }

    isPush.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isPushData(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rom.IS_PUSH_DATA already set");
    } else {
      filled.set(11);
    }

    isPushData.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rom.LIMB already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rom.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

    return this;
  }

  public Trace nBytes(final UnsignedByte b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rom.nBYTES already set");
    } else {
      filled.set(21);
    }

    nBytes.put(b.toByte());

    return this;
  }

  public Trace nBytesAcc(final UnsignedByte b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rom.nBYTES_ACC already set");
    } else {
      filled.set(22);
    }

    nBytesAcc.put(b.toByte());

    return this;
  }

  public Trace opcode(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rom.OPCODE already set");
    } else {
      filled.set(13);
    }

    opcode.put(b.toByte());

    return this;
  }

  public Trace paddedBytecodeByte(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rom.PADDED_BYTECODE_BYTE already set");
    } else {
      filled.set(14);
    }

    paddedBytecodeByte.put(b.toByte());

    return this;
  }

  public Trace programCounter(final long b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rom.PROGRAM_COUNTER already set");
    } else {
      filled.set(15);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rom.PROGRAM_COUNTER has invalid value (" + b + ")"); }
    programCounter.put((byte) (b >> 24));
    programCounter.put((byte) (b >> 16));
    programCounter.put((byte) (b >> 8));
    programCounter.put((byte) b);


    return this;
  }

  public Trace pushFunnelBit(final Boolean b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rom.PUSH_FUNNEL_BIT already set");
    } else {
      filled.set(16);
    }

    pushFunnelBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pushParameter(final UnsignedByte b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rom.PUSH_PARAMETER already set");
    } else {
      filled.set(17);
    }

    pushParameter.put(b.toByte());

    return this;
  }

  public Trace pushValueAcc(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rom.PUSH_VALUE_ACC already set");
    } else {
      filled.set(18);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rom.PUSH_VALUE_ACC has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { pushValueAcc.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { pushValueAcc.put(bs.get(j)); }

    return this;
  }

  public Trace pushValueHi(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rom.PUSH_VALUE_HI already set");
    } else {
      filled.set(19);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rom.PUSH_VALUE_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { pushValueHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { pushValueHi.put(bs.get(j)); }

    return this;
  }

  public Trace pushValueLo(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rom.PUSH_VALUE_LO already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rom.PUSH_VALUE_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { pushValueLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { pushValueLo.put(bs.get(j)); }

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
      throw new IllegalStateException("rom.IS_JUMPDEST has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rom.IS_PUSH has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rom.IS_PUSH_DATA has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rom.LIMB has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rom.nBYTES has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rom.nBYTES_ACC has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rom.OPCODE has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rom.PADDED_BYTECODE_BYTE has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rom.PROGRAM_COUNTER has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rom.PUSH_FUNNEL_BIT has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rom.PUSH_PARAMETER has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rom.PUSH_VALUE_ACC has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rom.PUSH_VALUE_HI has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rom.PUSH_VALUE_LO has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc.position(acc.position() + 16);
    }

    if (!filled.get(2)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 4);
    }

    if (!filled.get(3)) {
      codeFragmentIndexInfty.position(codeFragmentIndexInfty.position() + 4);
    }

    if (!filled.get(4)) {
      codeSize.position(codeSize.position() + 4);
    }

    if (!filled.get(1)) {
      codesizeReached.position(codesizeReached.position() + 1);
    }

    if (!filled.get(5)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(6)) {
      counterMax.position(counterMax.position() + 1);
    }

    if (!filled.get(7)) {
      counterPush.position(counterPush.position() + 1);
    }

    if (!filled.get(8)) {
      index.position(index.position() + 4);
    }

    if (!filled.get(9)) {
      isJumpdest.position(isJumpdest.position() + 1);
    }

    if (!filled.get(10)) {
      isPush.position(isPush.position() + 1);
    }

    if (!filled.get(11)) {
      isPushData.position(isPushData.position() + 1);
    }

    if (!filled.get(12)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(21)) {
      nBytes.position(nBytes.position() + 1);
    }

    if (!filled.get(22)) {
      nBytesAcc.position(nBytesAcc.position() + 1);
    }

    if (!filled.get(13)) {
      opcode.position(opcode.position() + 1);
    }

    if (!filled.get(14)) {
      paddedBytecodeByte.position(paddedBytecodeByte.position() + 1);
    }

    if (!filled.get(15)) {
      programCounter.position(programCounter.position() + 4);
    }

    if (!filled.get(16)) {
      pushFunnelBit.position(pushFunnelBit.position() + 1);
    }

    if (!filled.get(17)) {
      pushParameter.position(pushParameter.position() + 1);
    }

    if (!filled.get(18)) {
      pushValueAcc.position(pushValueAcc.position() + 16);
    }

    if (!filled.get(19)) {
      pushValueHi.position(pushValueHi.position() + 16);
    }

    if (!filled.get(20)) {
      pushValueLo.position(pushValueLo.position() + 16);
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
