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

package net.consensys.linea.zktracer.module.rlpaddr;

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
  public static final int MAX_CT_CREATE = 0x7;
  public static final int MAX_CT_CREATE2 = 0x5;

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
  private final MappedByteBuffer rawAddrHi;
  private final MappedByteBuffer recipe;
  private final MappedByteBuffer recipe1;
  private final MappedByteBuffer recipe2;
  private final MappedByteBuffer saltHi;
  private final MappedByteBuffer saltLo;
  private final MappedByteBuffer selectorKeccakRes;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer tinyNonZeroNonce;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("rlpaddr.ACC", 8, length));
      headers.add(new ColumnHeader("rlpaddr.ACC_BYTESIZE", 1, length));
      headers.add(new ColumnHeader("rlpaddr.ADDR_HI", 4, length));
      headers.add(new ColumnHeader("rlpaddr.ADDR_LO", 16, length));
      headers.add(new ColumnHeader("rlpaddr.BIT1", 1, length));
      headers.add(new ColumnHeader("rlpaddr.BIT_ACC", 1, length));
      headers.add(new ColumnHeader("rlpaddr.BYTE1", 1, length));
      headers.add(new ColumnHeader("rlpaddr.COUNTER", 1, length));
      headers.add(new ColumnHeader("rlpaddr.DEP_ADDR_HI", 4, length));
      headers.add(new ColumnHeader("rlpaddr.DEP_ADDR_LO", 16, length));
      headers.add(new ColumnHeader("rlpaddr.INDEX", 1, length));
      headers.add(new ColumnHeader("rlpaddr.KEC_HI", 16, length));
      headers.add(new ColumnHeader("rlpaddr.KEC_LO", 16, length));
      headers.add(new ColumnHeader("rlpaddr.LC", 1, length));
      headers.add(new ColumnHeader("rlpaddr.LIMB", 16, length));
      headers.add(new ColumnHeader("rlpaddr.nBYTES", 1, length));
      headers.add(new ColumnHeader("rlpaddr.NONCE", 8, length));
      headers.add(new ColumnHeader("rlpaddr.POWER", 16, length));
      headers.add(new ColumnHeader("rlpaddr.RAW_ADDR_HI", 16, length));
      headers.add(new ColumnHeader("rlpaddr.RECIPE", 1, length));
      headers.add(new ColumnHeader("rlpaddr.RECIPE_1", 1, length));
      headers.add(new ColumnHeader("rlpaddr.RECIPE_2", 1, length));
      headers.add(new ColumnHeader("rlpaddr.SALT_HI", 16, length));
      headers.add(new ColumnHeader("rlpaddr.SALT_LO", 16, length));
      headers.add(new ColumnHeader("rlpaddr.SELECTOR_KECCAK_RES", 1, length));
      headers.add(new ColumnHeader("rlpaddr.STAMP", 3, length));
      headers.add(new ColumnHeader("rlpaddr.TINY_NON_ZERO_NONCE", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
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
    this.rawAddrHi = buffers.get(18);
    this.recipe = buffers.get(19);
    this.recipe1 = buffers.get(20);
    this.recipe2 = buffers.get(21);
    this.saltHi = buffers.get(22);
    this.saltLo = buffers.get(23);
    this.selectorKeccakRes = buffers.get(24);
    this.stamp = buffers.get(25);
    this.tinyNonZeroNonce = buffers.get(26);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("rlpaddr.ACC already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("rlpaddr.ACC has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { acc.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc.put(bs.get(j)); }

    return this;
  }

  public Trace accBytesize(final UnsignedByte b) {
    if (filled.get(1)) {
      throw new IllegalStateException("rlpaddr.ACC_BYTESIZE already set");
    } else {
      filled.set(1);
    }

    accBytesize.put(b.toByte());

    return this;
  }

  public Trace addrHi(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("rlpaddr.ADDR_HI already set");
    } else {
      filled.set(2);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlpaddr.ADDR_HI has invalid value (" + b + ")"); }
    addrHi.put((byte) (b >> 24));
    addrHi.put((byte) (b >> 16));
    addrHi.put((byte) (b >> 8));
    addrHi.put((byte) b);


    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("rlpaddr.ADDR_LO already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.ADDR_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addrLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addrLo.put(bs.get(j)); }

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(4)) {
      throw new IllegalStateException("rlpaddr.BIT1 already set");
    } else {
      filled.set(4);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bitAcc(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("rlpaddr.BIT_ACC already set");
    } else {
      filled.set(5);
    }

    bitAcc.put(b.toByte());

    return this;
  }

  public Trace byte1(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("rlpaddr.BYTE1 already set");
    } else {
      filled.set(6);
    }

    byte1.put(b.toByte());

    return this;
  }

  public Trace counter(final UnsignedByte b) {
    if (filled.get(7)) {
      throw new IllegalStateException("rlpaddr.COUNTER already set");
    } else {
      filled.set(7);
    }

    counter.put(b.toByte());

    return this;
  }

  public Trace depAddrHi(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("rlpaddr.DEP_ADDR_HI already set");
    } else {
      filled.set(8);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("rlpaddr.DEP_ADDR_HI has invalid value (" + b + ")"); }
    depAddrHi.put((byte) (b >> 24));
    depAddrHi.put((byte) (b >> 16));
    depAddrHi.put((byte) (b >> 8));
    depAddrHi.put((byte) b);


    return this;
  }

  public Trace depAddrLo(final Bytes b) {
    if (filled.get(9)) {
      throw new IllegalStateException("rlpaddr.DEP_ADDR_LO already set");
    } else {
      filled.set(9);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.DEP_ADDR_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { depAddrLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { depAddrLo.put(bs.get(j)); }

    return this;
  }

  public Trace index(final UnsignedByte b) {
    if (filled.get(10)) {
      throw new IllegalStateException("rlpaddr.INDEX already set");
    } else {
      filled.set(10);
    }

    index.put(b.toByte());

    return this;
  }

  public Trace kecHi(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("rlpaddr.KEC_HI already set");
    } else {
      filled.set(11);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.KEC_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { kecHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { kecHi.put(bs.get(j)); }

    return this;
  }

  public Trace kecLo(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("rlpaddr.KEC_LO already set");
    } else {
      filled.set(12);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.KEC_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { kecLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { kecLo.put(bs.get(j)); }

    return this;
  }

  public Trace lc(final Boolean b) {
    if (filled.get(13)) {
      throw new IllegalStateException("rlpaddr.LC already set");
    } else {
      filled.set(13);
    }

    lc.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(14)) {
      throw new IllegalStateException("rlpaddr.LIMB already set");
    } else {
      filled.set(14);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

    return this;
  }

  public Trace nBytes(final UnsignedByte b) {
    if (filled.get(26)) {
      throw new IllegalStateException("rlpaddr.nBYTES already set");
    } else {
      filled.set(26);
    }

    nBytes.put(b.toByte());

    return this;
  }

  public Trace nonce(final Bytes b) {
    if (filled.get(15)) {
      throw new IllegalStateException("rlpaddr.NONCE already set");
    } else {
      filled.set(15);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("rlpaddr.NONCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { nonce.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { nonce.put(bs.get(j)); }

    return this;
  }

  public Trace power(final Bytes b) {
    if (filled.get(16)) {
      throw new IllegalStateException("rlpaddr.POWER already set");
    } else {
      filled.set(16);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.POWER has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { power.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { power.put(bs.get(j)); }

    return this;
  }

  public Trace rawAddrHi(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("rlpaddr.RAW_ADDR_HI already set");
    } else {
      filled.set(17);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.RAW_ADDR_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { rawAddrHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { rawAddrHi.put(bs.get(j)); }

    return this;
  }

  public Trace recipe(final UnsignedByte b) {
    if (filled.get(18)) {
      throw new IllegalStateException("rlpaddr.RECIPE already set");
    } else {
      filled.set(18);
    }

    recipe.put(b.toByte());

    return this;
  }

  public Trace recipe1(final Boolean b) {
    if (filled.get(19)) {
      throw new IllegalStateException("rlpaddr.RECIPE_1 already set");
    } else {
      filled.set(19);
    }

    recipe1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace recipe2(final Boolean b) {
    if (filled.get(20)) {
      throw new IllegalStateException("rlpaddr.RECIPE_2 already set");
    } else {
      filled.set(20);
    }

    recipe2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace saltHi(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("rlpaddr.SALT_HI already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.SALT_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { saltHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { saltHi.put(bs.get(j)); }

    return this;
  }

  public Trace saltLo(final Bytes b) {
    if (filled.get(22)) {
      throw new IllegalStateException("rlpaddr.SALT_LO already set");
    } else {
      filled.set(22);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("rlpaddr.SALT_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { saltLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { saltLo.put(bs.get(j)); }

    return this;
  }

  public Trace selectorKeccakRes(final Boolean b) {
    if (filled.get(23)) {
      throw new IllegalStateException("rlpaddr.SELECTOR_KECCAK_RES already set");
    } else {
      filled.set(23);
    }

    selectorKeccakRes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final long b) {
    if (filled.get(24)) {
      throw new IllegalStateException("rlpaddr.STAMP already set");
    } else {
      filled.set(24);
    }

    if(b >= 16777216L) { throw new IllegalArgumentException("rlpaddr.STAMP has invalid value (" + b + ")"); }
    stamp.put((byte) (b >> 16));
    stamp.put((byte) (b >> 8));
    stamp.put((byte) b);


    return this;
  }

  public Trace tinyNonZeroNonce(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("rlpaddr.TINY_NON_ZERO_NONCE already set");
    } else {
      filled.set(25);
    }

    tinyNonZeroNonce.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("rlpaddr.ACC has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("rlpaddr.ACC_BYTESIZE has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("rlpaddr.ADDR_HI has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("rlpaddr.ADDR_LO has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("rlpaddr.BIT1 has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("rlpaddr.BIT_ACC has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("rlpaddr.BYTE1 has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("rlpaddr.COUNTER has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("rlpaddr.DEP_ADDR_HI has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("rlpaddr.DEP_ADDR_LO has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("rlpaddr.INDEX has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("rlpaddr.KEC_HI has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("rlpaddr.KEC_LO has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("rlpaddr.LC has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("rlpaddr.LIMB has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("rlpaddr.nBYTES has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("rlpaddr.NONCE has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("rlpaddr.POWER has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("rlpaddr.RAW_ADDR_HI has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("rlpaddr.RECIPE has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("rlpaddr.RECIPE_1 has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("rlpaddr.RECIPE_2 has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("rlpaddr.SALT_HI has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("rlpaddr.SALT_LO has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("rlpaddr.SELECTOR_KECCAK_RES has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("rlpaddr.STAMP has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("rlpaddr.TINY_NON_ZERO_NONCE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc.position(acc.position() + 8);
    }

    if (!filled.get(1)) {
      accBytesize.position(accBytesize.position() + 1);
    }

    if (!filled.get(2)) {
      addrHi.position(addrHi.position() + 4);
    }

    if (!filled.get(3)) {
      addrLo.position(addrLo.position() + 16);
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
      counter.position(counter.position() + 1);
    }

    if (!filled.get(8)) {
      depAddrHi.position(depAddrHi.position() + 4);
    }

    if (!filled.get(9)) {
      depAddrLo.position(depAddrLo.position() + 16);
    }

    if (!filled.get(10)) {
      index.position(index.position() + 1);
    }

    if (!filled.get(11)) {
      kecHi.position(kecHi.position() + 16);
    }

    if (!filled.get(12)) {
      kecLo.position(kecLo.position() + 16);
    }

    if (!filled.get(13)) {
      lc.position(lc.position() + 1);
    }

    if (!filled.get(14)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(26)) {
      nBytes.position(nBytes.position() + 1);
    }

    if (!filled.get(15)) {
      nonce.position(nonce.position() + 8);
    }

    if (!filled.get(16)) {
      power.position(power.position() + 16);
    }

    if (!filled.get(17)) {
      rawAddrHi.position(rawAddrHi.position() + 16);
    }

    if (!filled.get(18)) {
      recipe.position(recipe.position() + 1);
    }

    if (!filled.get(19)) {
      recipe1.position(recipe1.position() + 1);
    }

    if (!filled.get(20)) {
      recipe2.position(recipe2.position() + 1);
    }

    if (!filled.get(21)) {
      saltHi.position(saltHi.position() + 16);
    }

    if (!filled.get(22)) {
      saltLo.position(saltLo.position() + 16);
    }

    if (!filled.get(23)) {
      selectorKeccakRes.position(selectorKeccakRes.position() + 1);
    }

    if (!filled.get(24)) {
      stamp.position(stamp.position() + 3);
    }

    if (!filled.get(25)) {
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
