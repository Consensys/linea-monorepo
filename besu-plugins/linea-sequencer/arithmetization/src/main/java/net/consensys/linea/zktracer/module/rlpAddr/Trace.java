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

package net.consensys.linea.zktracer.module.rlpAddr;

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
  static final int CREATE2_SHIFT = 255;
  static final int G_TXDATA_NONZERO = 16;
  static final int G_TXDATA_ZERO = 4;
  static final int INT_LONG = 183;
  static final int INT_SHORT = 128;
  static final int LIST_LONG = 247;
  static final int LIST_SHORT = 192;
  static final int LLARGE = 16;
  static final int LLARGEMO = 15;
  static final int RLPADDR_CONST_RECIPE_1 = 1;
  static final int RLPADDR_CONST_RECIPE_2 = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_ADDR = 53;
  static final int RLPRECEIPT_SUBPHASE_ID_CUMUL_GAS = 3;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_LIMB = 77;
  static final int RLPRECEIPT_SUBPHASE_ID_DATA_SIZE = 83;
  static final int RLPRECEIPT_SUBPHASE_ID_NO_LOG_ENTRY = 11;
  static final int RLPRECEIPT_SUBPHASE_ID_STATUS_CODE = 2;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_BASE = 65;
  static final int RLPRECEIPT_SUBPHASE_ID_TOPIC_DELTA = 96;
  static final int RLPRECEIPT_SUBPHASE_ID_TYPE = 7;

  private final BitSet filled = new BitSet();
  private int currentLine = 0;

  private final MappedByteBuffer acc;
  private final MappedByteBuffer accBytesize;
  private final MappedByteBuffer addrHi;
  private final MappedByteBuffer addrLo;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bitAcc;
  private final MappedByteBuffer byte1;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer depAddrHi;
  private final MappedByteBuffer depAddrLo;
  private final MappedByteBuffer index;
  private final MappedByteBuffer kecHi;
  private final MappedByteBuffer kecLo;
  private final MappedByteBuffer lc;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer nBytes;
  private final MappedByteBuffer nonce;
  private final MappedByteBuffer power;
  private final MappedByteBuffer recipe;
  private final MappedByteBuffer recipe1;
  private final MappedByteBuffer recipe2;
  private final MappedByteBuffer saltHi;
  private final MappedByteBuffer saltLo;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer tinyNonZeroNonce;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("rlpAddr.ACC", 32, length),
        new ColumnHeader("rlpAddr.ACC_BYTESIZE", 32, length),
        new ColumnHeader("rlpAddr.ADDR_HI", 32, length),
        new ColumnHeader("rlpAddr.ADDR_LO", 32, length),
        new ColumnHeader("rlpAddr.BIT1", 1, length),
        new ColumnHeader("rlpAddr.BIT_ACC", 1, length),
        new ColumnHeader("rlpAddr.BYTE1", 1, length),
        new ColumnHeader("rlpAddr.COUNTER", 32, length),
        new ColumnHeader("rlpAddr.DEP_ADDR_HI", 32, length),
        new ColumnHeader("rlpAddr.DEP_ADDR_LO", 32, length),
        new ColumnHeader("rlpAddr.INDEX", 32, length),
        new ColumnHeader("rlpAddr.KEC_HI", 32, length),
        new ColumnHeader("rlpAddr.KEC_LO", 32, length),
        new ColumnHeader("rlpAddr.LC", 1, length),
        new ColumnHeader("rlpAddr.LIMB", 32, length),
        new ColumnHeader("rlpAddr.nBYTES", 32, length),
        new ColumnHeader("rlpAddr.NONCE", 32, length),
        new ColumnHeader("rlpAddr.POWER", 32, length),
        new ColumnHeader("rlpAddr.RECIPE", 32, length),
        new ColumnHeader("rlpAddr.RECIPE_1", 1, length),
        new ColumnHeader("rlpAddr.RECIPE_2", 1, length),
        new ColumnHeader("rlpAddr.SALT_HI", 32, length),
        new ColumnHeader("rlpAddr.SALT_LO", 32, length),
        new ColumnHeader("rlpAddr.STAMP", 32, length),
        new ColumnHeader("rlpAddr.TINY_NON_ZERO_NONCE", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.acc = buffers.get(0);
    this.accBytesize = buffers.get(1);
    this.addrHi = buffers.get(2);
    this.addrLo = buffers.get(3);
    this.bit1 = buffers.get(4);
    this.bitAcc = buffers.get(5);
    this.byte1 = buffers.get(6);
    this.counter = buffers.get(7);
    this.depAddrHi = buffers.get(8);
    this.depAddrLo = buffers.get(9);
    this.index = buffers.get(10);
    this.kecHi = buffers.get(11);
    this.kecLo = buffers.get(12);
    this.lc = buffers.get(13);
    this.limb = buffers.get(14);
    this.nBytes = buffers.get(15);
    this.nonce = buffers.get(16);
    this.power = buffers.get(17);
    this.recipe = buffers.get(18);
    this.recipe1 = buffers.get(19);
    this.recipe2 = buffers.get(20);
    this.saltHi = buffers.get(21);
    this.saltLo = buffers.get(22);
    this.stamp = buffers.get(23);
    this.tinyNonZeroNonce = buffers.get(24);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlpAddr.ACC already set");
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

  public Trace accBytesize(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlpAddr.ACC_BYTESIZE already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accBytesize.put((byte) 0);
    }
    accBytesize.put(b.toArrayUnsafe());

    return this;
  }

  public Trace addrHi(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlpAddr.ADDR_HI already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrHi.put((byte) 0);
    }
    addrHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlpAddr.ADDR_LO already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrLo.put((byte) 0);
    }
    addrLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlpAddr.BIT1 already set");
    } else {
      filled.set(4);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlpAddr.BIT_ACC already set");
    } else {
      filled.set(5);
    }

    bitAcc.put(b.toByte());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlpAddr.BYTE1 already set");
    } else {
      filled.set(6);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace counter(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlpAddr.COUNTER already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      counter.put((byte) 0);
    }
    counter.put(b.toArrayUnsafe());

    return this;
  }

  public Trace depAddrHi(final Bytes b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlpAddr.DEP_ADDR_HI already set");
    } else {
      filled.set(8);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      depAddrHi.put((byte) 0);
    }
    depAddrHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace depAddrLo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlpAddr.DEP_ADDR_LO already set");
    } else {
      filled.set(9);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      depAddrLo.put((byte) 0);
    }
    depAddrLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace index(final Bytes b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlpAddr.INDEX already set");
    } else {
      filled.set(10);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      index.put((byte) 0);
    }
    index.put(b.toArrayUnsafe());

    return this;
  }

  public Trace kecHi(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlpAddr.KEC_HI already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      kecHi.put((byte) 0);
    }
    kecHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace kecLo(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlpAddr.KEC_LO already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      kecLo.put((byte) 0);
    }
    kecLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace lc(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlpAddr.LC already set");
    } else {
      filled.set(13);
    }

    lc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlpAddr.LIMB already set");
    } else {
      filled.set(14);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      limb.put((byte) 0);
    }
    limb.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nBytes(final Bytes b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlpAddr.nBYTES already set");
    } else {
      filled.set(24);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nBytes.put((byte) 0);
    }
    nBytes.put(b.toArrayUnsafe());

    return this;
  }

  public Trace nonce(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlpAddr.NONCE already set");
    } else {
      filled.set(15);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      nonce.put((byte) 0);
    }
    nonce.put(b.toArrayUnsafe());

    return this;
  }

  public Trace power(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlpAddr.POWER already set");
    } else {
      filled.set(16);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      power.put((byte) 0);
    }
    power.put(b.toArrayUnsafe());

    return this;
  }

  public Trace recipe(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlpAddr.RECIPE already set");
    } else {
      filled.set(17);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      recipe.put((byte) 0);
    }
    recipe.put(b.toArrayUnsafe());

    return this;
  }

  public Trace recipe1(final Boolean b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlpAddr.RECIPE_1 already set");
    } else {
      filled.set(18);
    }

    recipe1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace recipe2(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlpAddr.RECIPE_2 already set");
    } else {
      filled.set(19);
    }

    recipe2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace saltHi(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlpAddr.SALT_HI already set");
    } else {
      filled.set(20);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      saltHi.put((byte) 0);
    }
    saltHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace saltLo(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlpAddr.SALT_LO already set");
    } else {
      filled.set(21);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      saltLo.put((byte) 0);
    }
    saltLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace stamp(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlpAddr.STAMP already set");
    } else {
      filled.set(22);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace tinyNonZeroNonce(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlpAddr.TINY_NON_ZERO_NONCE already set");
    } else {
      filled.set(23);
    }

    tinyNonZeroNonce.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlpAddr.ACC has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlpAddr.ACC_BYTESIZE has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlpAddr.ADDR_HI has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlpAddr.ADDR_LO has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlpAddr.BIT1 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlpAddr.BIT_ACC has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlpAddr.BYTE1 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlpAddr.COUNTER has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlpAddr.DEP_ADDR_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlpAddr.DEP_ADDR_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlpAddr.INDEX has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlpAddr.KEC_HI has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlpAddr.KEC_LO has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlpAddr.LC has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlpAddr.LIMB has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlpAddr.nBYTES has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlpAddr.NONCE has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlpAddr.POWER has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlpAddr.RECIPE has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlpAddr.RECIPE_1 has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlpAddr.RECIPE_2 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlpAddr.SALT_HI has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlpAddr.SALT_LO has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlpAddr.STAMP has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlpAddr.TINY_NON_ZERO_NONCE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc.position(acc.position() + 32);
    }

    if (!filled.get(1)) {
      accBytesize.position(accBytesize.position() + 32);
    }

    if (!filled.get(2)) {
      addrHi.position(addrHi.position() + 32);
    }

    if (!filled.get(3)) {
      addrLo.position(addrLo.position() + 32);
    }

    if (!filled.get(4)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(5)) {
      bitAcc.position(bitAcc.position() + 1);
    }

    if (!filled.get(6)) {
      byte1.position(byte1.position() + 1);
    }

    if (!filled.get(7)) {
      counter.position(counter.position() + 32);
    }

    if (!filled.get(8)) {
      depAddrHi.position(depAddrHi.position() + 32);
    }

    if (!filled.get(9)) {
      depAddrLo.position(depAddrLo.position() + 32);
    }

    if (!filled.get(10)) {
      index.position(index.position() + 32);
    }

    if (!filled.get(11)) {
      kecHi.position(kecHi.position() + 32);
    }

    if (!filled.get(12)) {
      kecLo.position(kecLo.position() + 32);
    }

    if (!filled.get(13)) {
      lc.position(lc.position() + 1);
    }

    if (!filled.get(14)) {
      limb.position(limb.position() + 32);
    }

    if (!filled.get(24)) {
      nBytes.position(nBytes.position() + 32);
    }

    if (!filled.get(15)) {
      nonce.position(nonce.position() + 32);
    }

    if (!filled.get(16)) {
      power.position(power.position() + 32);
    }

    if (!filled.get(17)) {
      recipe.position(recipe.position() + 32);
    }

    if (!filled.get(18)) {
      recipe1.position(recipe1.position() + 1);
    }

    if (!filled.get(19)) {
      recipe2.position(recipe2.position() + 1);
    }

    if (!filled.get(20)) {
      saltHi.position(saltHi.position() + 32);
    }

    if (!filled.get(21)) {
      saltLo.position(saltLo.position() + 32);
    }

    if (!filled.get(22)) {
      stamp.position(stamp.position() + 32);
    }

    if (!filled.get(23)) {
      tinyNonZeroNonce.position(tinyNonZeroNonce.position() + 1);
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
