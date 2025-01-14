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

package net.consensys.linea.zktracer.module.mmio;

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

  private final MappedByteBuffer acc1;
  private final MappedByteBuffer acc2;
  private final MappedByteBuffer acc3;
  private final MappedByteBuffer acc4;
  private final MappedByteBuffer accA;
  private final MappedByteBuffer accB;
  private final MappedByteBuffer accC;
  private final MappedByteBuffer accLimb;
  private final MappedByteBuffer bit1;
  private final MappedByteBuffer bit2;
  private final MappedByteBuffer bit3;
  private final MappedByteBuffer bit4;
  private final MappedByteBuffer bit5;
  private final MappedByteBuffer byteA;
  private final MappedByteBuffer byteB;
  private final MappedByteBuffer byteC;
  private final MappedByteBuffer byteLimb;
  private final MappedByteBuffer cnA;
  private final MappedByteBuffer cnB;
  private final MappedByteBuffer cnC;
  private final MappedByteBuffer contextSource;
  private final MappedByteBuffer contextTarget;
  private final MappedByteBuffer counter;
  private final MappedByteBuffer exoId;
  private final MappedByteBuffer exoIsBlakemodexp;
  private final MappedByteBuffer exoIsEcdata;
  private final MappedByteBuffer exoIsKec;
  private final MappedByteBuffer exoIsLog;
  private final MappedByteBuffer exoIsRipsha;
  private final MappedByteBuffer exoIsRom;
  private final MappedByteBuffer exoIsTxcd;
  private final MappedByteBuffer exoSum;
  private final MappedByteBuffer fast;
  private final MappedByteBuffer indexA;
  private final MappedByteBuffer indexB;
  private final MappedByteBuffer indexC;
  private final MappedByteBuffer indexX;
  private final MappedByteBuffer isLimbToRamOneTarget;
  private final MappedByteBuffer isLimbToRamTransplant;
  private final MappedByteBuffer isLimbToRamTwoTarget;
  private final MappedByteBuffer isLimbVanishes;
  private final MappedByteBuffer isRamExcision;
  private final MappedByteBuffer isRamToLimbOneSource;
  private final MappedByteBuffer isRamToLimbTransplant;
  private final MappedByteBuffer isRamToLimbTwoSource;
  private final MappedByteBuffer isRamToRamPartial;
  private final MappedByteBuffer isRamToRamTransplant;
  private final MappedByteBuffer isRamToRamTwoSource;
  private final MappedByteBuffer isRamToRamTwoTarget;
  private final MappedByteBuffer isRamVanishes;
  private final MappedByteBuffer kecId;
  private final MappedByteBuffer limb;
  private final MappedByteBuffer mmioInstruction;
  private final MappedByteBuffer mmioStamp;
  private final MappedByteBuffer phase;
  private final MappedByteBuffer pow2561;
  private final MappedByteBuffer pow2562;
  private final MappedByteBuffer size;
  private final MappedByteBuffer slow;
  private final MappedByteBuffer sourceByteOffset;
  private final MappedByteBuffer sourceLimbOffset;
  private final MappedByteBuffer successBit;
  private final MappedByteBuffer targetByteOffset;
  private final MappedByteBuffer targetLimbOffset;
  private final MappedByteBuffer totalSize;
  private final MappedByteBuffer valA;
  private final MappedByteBuffer valANew;
  private final MappedByteBuffer valB;
  private final MappedByteBuffer valBNew;
  private final MappedByteBuffer valC;
  private final MappedByteBuffer valCNew;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("mmio.ACC_1", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_2", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_3", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_4", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_A", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_B", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_C", 16, length));
      headers.add(new ColumnHeader("mmio.ACC_LIMB", 16, length));
      headers.add(new ColumnHeader("mmio.BIT_1", 1, length));
      headers.add(new ColumnHeader("mmio.BIT_2", 1, length));
      headers.add(new ColumnHeader("mmio.BIT_3", 1, length));
      headers.add(new ColumnHeader("mmio.BIT_4", 1, length));
      headers.add(new ColumnHeader("mmio.BIT_5", 1, length));
      headers.add(new ColumnHeader("mmio.BYTE_A", 1, length));
      headers.add(new ColumnHeader("mmio.BYTE_B", 1, length));
      headers.add(new ColumnHeader("mmio.BYTE_C", 1, length));
      headers.add(new ColumnHeader("mmio.BYTE_LIMB", 1, length));
      headers.add(new ColumnHeader("mmio.CN_A", 8, length));
      headers.add(new ColumnHeader("mmio.CN_B", 8, length));
      headers.add(new ColumnHeader("mmio.CN_C", 8, length));
      headers.add(new ColumnHeader("mmio.CONTEXT_SOURCE", 8, length));
      headers.add(new ColumnHeader("mmio.CONTEXT_TARGET", 8, length));
      headers.add(new ColumnHeader("mmio.COUNTER", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_ID", 4, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_BLAKEMODEXP", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_ECDATA", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_KEC", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_LOG", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_RIPSHA", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_ROM", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_IS_TXCD", 1, length));
      headers.add(new ColumnHeader("mmio.EXO_SUM", 4, length));
      headers.add(new ColumnHeader("mmio.FAST", 1, length));
      headers.add(new ColumnHeader("mmio.INDEX_A", 8, length));
      headers.add(new ColumnHeader("mmio.INDEX_B", 8, length));
      headers.add(new ColumnHeader("mmio.INDEX_C", 8, length));
      headers.add(new ColumnHeader("mmio.INDEX_X", 8, length));
      headers.add(new ColumnHeader("mmio.IS_LIMB_TO_RAM_ONE_TARGET", 1, length));
      headers.add(new ColumnHeader("mmio.IS_LIMB_TO_RAM_TRANSPLANT", 1, length));
      headers.add(new ColumnHeader("mmio.IS_LIMB_TO_RAM_TWO_TARGET", 1, length));
      headers.add(new ColumnHeader("mmio.IS_LIMB_VANISHES", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_EXCISION", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_LIMB_ONE_SOURCE", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_LIMB_TRANSPLANT", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_LIMB_TWO_SOURCE", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_RAM_PARTIAL", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_RAM_TRANSPLANT", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_RAM_TWO_SOURCE", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_TO_RAM_TWO_TARGET", 1, length));
      headers.add(new ColumnHeader("mmio.IS_RAM_VANISHES", 1, length));
      headers.add(new ColumnHeader("mmio.KEC_ID", 4, length));
      headers.add(new ColumnHeader("mmio.LIMB", 16, length));
      headers.add(new ColumnHeader("mmio.MMIO_INSTRUCTION", 2, length));
      headers.add(new ColumnHeader("mmio.MMIO_STAMP", 4, length));
      headers.add(new ColumnHeader("mmio.PHASE", 4, length));
      headers.add(new ColumnHeader("mmio.POW_256_1", 16, length));
      headers.add(new ColumnHeader("mmio.POW_256_2", 16, length));
      headers.add(new ColumnHeader("mmio.SIZE", 8, length));
      headers.add(new ColumnHeader("mmio.SLOW", 1, length));
      headers.add(new ColumnHeader("mmio.SOURCE_BYTE_OFFSET", 1, length));
      headers.add(new ColumnHeader("mmio.SOURCE_LIMB_OFFSET", 8, length));
      headers.add(new ColumnHeader("mmio.SUCCESS_BIT", 1, length));
      headers.add(new ColumnHeader("mmio.TARGET_BYTE_OFFSET", 1, length));
      headers.add(new ColumnHeader("mmio.TARGET_LIMB_OFFSET", 8, length));
      headers.add(new ColumnHeader("mmio.TOTAL_SIZE", 8, length));
      headers.add(new ColumnHeader("mmio.VAL_A", 16, length));
      headers.add(new ColumnHeader("mmio.VAL_A_NEW", 16, length));
      headers.add(new ColumnHeader("mmio.VAL_B", 16, length));
      headers.add(new ColumnHeader("mmio.VAL_B_NEW", 16, length));
      headers.add(new ColumnHeader("mmio.VAL_C", 16, length));
      headers.add(new ColumnHeader("mmio.VAL_C_NEW", 16, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.acc1 = buffers.get(0);
    this.acc2 = buffers.get(1);
    this.acc3 = buffers.get(2);
    this.acc4 = buffers.get(3);
    this.accA = buffers.get(4);
    this.accB = buffers.get(5);
    this.accC = buffers.get(6);
    this.accLimb = buffers.get(7);
    this.bit1 = buffers.get(8);
    this.bit2 = buffers.get(9);
    this.bit3 = buffers.get(10);
    this.bit4 = buffers.get(11);
    this.bit5 = buffers.get(12);
    this.byteA = buffers.get(13);
    this.byteB = buffers.get(14);
    this.byteC = buffers.get(15);
    this.byteLimb = buffers.get(16);
    this.cnA = buffers.get(17);
    this.cnB = buffers.get(18);
    this.cnC = buffers.get(19);
    this.contextSource = buffers.get(20);
    this.contextTarget = buffers.get(21);
    this.counter = buffers.get(22);
    this.exoId = buffers.get(23);
    this.exoIsBlakemodexp = buffers.get(24);
    this.exoIsEcdata = buffers.get(25);
    this.exoIsKec = buffers.get(26);
    this.exoIsLog = buffers.get(27);
    this.exoIsRipsha = buffers.get(28);
    this.exoIsRom = buffers.get(29);
    this.exoIsTxcd = buffers.get(30);
    this.exoSum = buffers.get(31);
    this.fast = buffers.get(32);
    this.indexA = buffers.get(33);
    this.indexB = buffers.get(34);
    this.indexC = buffers.get(35);
    this.indexX = buffers.get(36);
    this.isLimbToRamOneTarget = buffers.get(37);
    this.isLimbToRamTransplant = buffers.get(38);
    this.isLimbToRamTwoTarget = buffers.get(39);
    this.isLimbVanishes = buffers.get(40);
    this.isRamExcision = buffers.get(41);
    this.isRamToLimbOneSource = buffers.get(42);
    this.isRamToLimbTransplant = buffers.get(43);
    this.isRamToLimbTwoSource = buffers.get(44);
    this.isRamToRamPartial = buffers.get(45);
    this.isRamToRamTransplant = buffers.get(46);
    this.isRamToRamTwoSource = buffers.get(47);
    this.isRamToRamTwoTarget = buffers.get(48);
    this.isRamVanishes = buffers.get(49);
    this.kecId = buffers.get(50);
    this.limb = buffers.get(51);
    this.mmioInstruction = buffers.get(52);
    this.mmioStamp = buffers.get(53);
    this.phase = buffers.get(54);
    this.pow2561 = buffers.get(55);
    this.pow2562 = buffers.get(56);
    this.size = buffers.get(57);
    this.slow = buffers.get(58);
    this.sourceByteOffset = buffers.get(59);
    this.sourceLimbOffset = buffers.get(60);
    this.successBit = buffers.get(61);
    this.targetByteOffset = buffers.get(62);
    this.targetLimbOffset = buffers.get(63);
    this.totalSize = buffers.get(64);
    this.valA = buffers.get(65);
    this.valANew = buffers.get(66);
    this.valB = buffers.get(67);
    this.valBNew = buffers.get(68);
    this.valC = buffers.get(69);
    this.valCNew = buffers.get(70);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace acc1(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("mmio.ACC_1 already set");
    } else {
      filled.set(0);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc1.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc1.put(bs.get(j)); }

    return this;
  }

  public Trace acc2(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("mmio.ACC_2 already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc2.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc2.put(bs.get(j)); }

    return this;
  }

  public Trace acc3(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("mmio.ACC_3 already set");
    } else {
      filled.set(2);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_3 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc3.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc3.put(bs.get(j)); }

    return this;
  }

  public Trace acc4(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("mmio.ACC_4 already set");
    } else {
      filled.set(3);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_4 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { acc4.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { acc4.put(bs.get(j)); }

    return this;
  }

  public Trace accA(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("mmio.ACC_A already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_A has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { accA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accA.put(bs.get(j)); }

    return this;
  }

  public Trace accB(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("mmio.ACC_B already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_B has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { accB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accB.put(bs.get(j)); }

    return this;
  }

  public Trace accC(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("mmio.ACC_C already set");
    } else {
      filled.set(6);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_C has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { accC.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accC.put(bs.get(j)); }

    return this;
  }

  public Trace accLimb(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("mmio.ACC_LIMB already set");
    } else {
      filled.set(7);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.ACC_LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { accLimb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { accLimb.put(bs.get(j)); }

    return this;
  }

  public Trace bit1(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("mmio.BIT_1 already set");
    } else {
      filled.set(8);
    }

    bit1.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit2(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("mmio.BIT_2 already set");
    } else {
      filled.set(9);
    }

    bit2.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit3(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("mmio.BIT_3 already set");
    } else {
      filled.set(10);
    }

    bit3.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit4(final Boolean b) {
    if (filled.get(11)) {
      throw new IllegalStateException("mmio.BIT_4 already set");
    } else {
      filled.set(11);
    }

    bit4.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace bit5(final Boolean b) {
    if (filled.get(12)) {
      throw new IllegalStateException("mmio.BIT_5 already set");
    } else {
      filled.set(12);
    }

    bit5.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace byteA(final UnsignedByte b) {
    if (filled.get(13)) {
      throw new IllegalStateException("mmio.BYTE_A already set");
    } else {
      filled.set(13);
    }

    byteA.put(b.toByte());

    return this;
  }

  public Trace byteB(final UnsignedByte b) {
    if (filled.get(14)) {
      throw new IllegalStateException("mmio.BYTE_B already set");
    } else {
      filled.set(14);
    }

    byteB.put(b.toByte());

    return this;
  }

  public Trace byteC(final UnsignedByte b) {
    if (filled.get(15)) {
      throw new IllegalStateException("mmio.BYTE_C already set");
    } else {
      filled.set(15);
    }

    byteC.put(b.toByte());

    return this;
  }

  public Trace byteLimb(final UnsignedByte b) {
    if (filled.get(16)) {
      throw new IllegalStateException("mmio.BYTE_LIMB already set");
    } else {
      filled.set(16);
    }

    byteLimb.put(b.toByte());

    return this;
  }

  public Trace cnA(final Bytes b) {
    if (filled.get(17)) {
      throw new IllegalStateException("mmio.CN_A already set");
    } else {
      filled.set(17);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.CN_A has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { cnA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { cnA.put(bs.get(j)); }

    return this;
  }

  public Trace cnB(final Bytes b) {
    if (filled.get(18)) {
      throw new IllegalStateException("mmio.CN_B already set");
    } else {
      filled.set(18);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.CN_B has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { cnB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { cnB.put(bs.get(j)); }

    return this;
  }

  public Trace cnC(final Bytes b) {
    if (filled.get(19)) {
      throw new IllegalStateException("mmio.CN_C already set");
    } else {
      filled.set(19);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.CN_C has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { cnC.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { cnC.put(bs.get(j)); }

    return this;
  }

  public Trace contextSource(final Bytes b) {
    if (filled.get(20)) {
      throw new IllegalStateException("mmio.CONTEXT_SOURCE already set");
    } else {
      filled.set(20);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.CONTEXT_SOURCE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { contextSource.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { contextSource.put(bs.get(j)); }

    return this;
  }

  public Trace contextTarget(final Bytes b) {
    if (filled.get(21)) {
      throw new IllegalStateException("mmio.CONTEXT_TARGET already set");
    } else {
      filled.set(21);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.CONTEXT_TARGET has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { contextTarget.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { contextTarget.put(bs.get(j)); }

    return this;
  }

  public Trace counter(final long b) {
    if (filled.get(22)) {
      throw new IllegalStateException("mmio.COUNTER already set");
    } else {
      filled.set(22);
    }

    if(b >= 256L) { throw new IllegalArgumentException("mmio.COUNTER has invalid value (" + b + ")"); }
    counter.put((byte) b);


    return this;
  }

  public Trace exoId(final long b) {
    if (filled.get(23)) {
      throw new IllegalStateException("mmio.EXO_ID already set");
    } else {
      filled.set(23);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmio.EXO_ID has invalid value (" + b + ")"); }
    exoId.put((byte) (b >> 24));
    exoId.put((byte) (b >> 16));
    exoId.put((byte) (b >> 8));
    exoId.put((byte) b);


    return this;
  }

  public Trace exoIsBlakemodexp(final Boolean b) {
    if (filled.get(24)) {
      throw new IllegalStateException("mmio.EXO_IS_BLAKEMODEXP already set");
    } else {
      filled.set(24);
    }

    exoIsBlakemodexp.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsEcdata(final Boolean b) {
    if (filled.get(25)) {
      throw new IllegalStateException("mmio.EXO_IS_ECDATA already set");
    } else {
      filled.set(25);
    }

    exoIsEcdata.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsKec(final Boolean b) {
    if (filled.get(26)) {
      throw new IllegalStateException("mmio.EXO_IS_KEC already set");
    } else {
      filled.set(26);
    }

    exoIsKec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsLog(final Boolean b) {
    if (filled.get(27)) {
      throw new IllegalStateException("mmio.EXO_IS_LOG already set");
    } else {
      filled.set(27);
    }

    exoIsLog.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsRipsha(final Boolean b) {
    if (filled.get(28)) {
      throw new IllegalStateException("mmio.EXO_IS_RIPSHA already set");
    } else {
      filled.set(28);
    }

    exoIsRipsha.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsRom(final Boolean b) {
    if (filled.get(29)) {
      throw new IllegalStateException("mmio.EXO_IS_ROM already set");
    } else {
      filled.set(29);
    }

    exoIsRom.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoIsTxcd(final Boolean b) {
    if (filled.get(30)) {
      throw new IllegalStateException("mmio.EXO_IS_TXCD already set");
    } else {
      filled.set(30);
    }

    exoIsTxcd.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace exoSum(final long b) {
    if (filled.get(31)) {
      throw new IllegalStateException("mmio.EXO_SUM already set");
    } else {
      filled.set(31);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmio.EXO_SUM has invalid value (" + b + ")"); }
    exoSum.put((byte) (b >> 24));
    exoSum.put((byte) (b >> 16));
    exoSum.put((byte) (b >> 8));
    exoSum.put((byte) b);


    return this;
  }

  public Trace fast(final Boolean b) {
    if (filled.get(32)) {
      throw new IllegalStateException("mmio.FAST already set");
    } else {
      filled.set(32);
    }

    fast.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace indexA(final Bytes b) {
    if (filled.get(33)) {
      throw new IllegalStateException("mmio.INDEX_A already set");
    } else {
      filled.set(33);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.INDEX_A has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { indexA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { indexA.put(bs.get(j)); }

    return this;
  }

  public Trace indexB(final Bytes b) {
    if (filled.get(34)) {
      throw new IllegalStateException("mmio.INDEX_B already set");
    } else {
      filled.set(34);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.INDEX_B has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { indexB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { indexB.put(bs.get(j)); }

    return this;
  }

  public Trace indexC(final Bytes b) {
    if (filled.get(35)) {
      throw new IllegalStateException("mmio.INDEX_C already set");
    } else {
      filled.set(35);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.INDEX_C has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { indexC.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { indexC.put(bs.get(j)); }

    return this;
  }

  public Trace indexX(final Bytes b) {
    if (filled.get(36)) {
      throw new IllegalStateException("mmio.INDEX_X already set");
    } else {
      filled.set(36);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.INDEX_X has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { indexX.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { indexX.put(bs.get(j)); }

    return this;
  }

  public Trace isLimbToRamOneTarget(final Boolean b) {
    if (filled.get(37)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_ONE_TARGET already set");
    } else {
      filled.set(37);
    }

    isLimbToRamOneTarget.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLimbToRamTransplant(final Boolean b) {
    if (filled.get(38)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TRANSPLANT already set");
    } else {
      filled.set(38);
    }

    isLimbToRamTransplant.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLimbToRamTwoTarget(final Boolean b) {
    if (filled.get(39)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TWO_TARGET already set");
    } else {
      filled.set(39);
    }

    isLimbToRamTwoTarget.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isLimbVanishes(final Boolean b) {
    if (filled.get(40)) {
      throw new IllegalStateException("mmio.IS_LIMB_VANISHES already set");
    } else {
      filled.set(40);
    }

    isLimbVanishes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamExcision(final Boolean b) {
    if (filled.get(41)) {
      throw new IllegalStateException("mmio.IS_RAM_EXCISION already set");
    } else {
      filled.set(41);
    }

    isRamExcision.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToLimbOneSource(final Boolean b) {
    if (filled.get(42)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_ONE_SOURCE already set");
    } else {
      filled.set(42);
    }

    isRamToLimbOneSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToLimbTransplant(final Boolean b) {
    if (filled.get(43)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TRANSPLANT already set");
    } else {
      filled.set(43);
    }

    isRamToLimbTransplant.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToLimbTwoSource(final Boolean b) {
    if (filled.get(44)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TWO_SOURCE already set");
    } else {
      filled.set(44);
    }

    isRamToLimbTwoSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamPartial(final Boolean b) {
    if (filled.get(45)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_PARTIAL already set");
    } else {
      filled.set(45);
    }

    isRamToRamPartial.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamTransplant(final Boolean b) {
    if (filled.get(46)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TRANSPLANT already set");
    } else {
      filled.set(46);
    }

    isRamToRamTransplant.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamTwoSource(final Boolean b) {
    if (filled.get(47)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_SOURCE already set");
    } else {
      filled.set(47);
    }

    isRamToRamTwoSource.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamToRamTwoTarget(final Boolean b) {
    if (filled.get(48)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_TARGET already set");
    } else {
      filled.set(48);
    }

    isRamToRamTwoTarget.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace isRamVanishes(final Boolean b) {
    if (filled.get(49)) {
      throw new IllegalStateException("mmio.IS_RAM_VANISHES already set");
    } else {
      filled.set(49);
    }

    isRamVanishes.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace kecId(final long b) {
    if (filled.get(50)) {
      throw new IllegalStateException("mmio.KEC_ID already set");
    } else {
      filled.set(50);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmio.KEC_ID has invalid value (" + b + ")"); }
    kecId.put((byte) (b >> 24));
    kecId.put((byte) (b >> 16));
    kecId.put((byte) (b >> 8));
    kecId.put((byte) b);


    return this;
  }

  public Trace limb(final Bytes b) {
    if (filled.get(51)) {
      throw new IllegalStateException("mmio.LIMB already set");
    } else {
      filled.set(51);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.LIMB has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { limb.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { limb.put(bs.get(j)); }

    return this;
  }

  public Trace mmioInstruction(final long b) {
    if (filled.get(52)) {
      throw new IllegalStateException("mmio.MMIO_INSTRUCTION already set");
    } else {
      filled.set(52);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("mmio.MMIO_INSTRUCTION has invalid value (" + b + ")"); }
    mmioInstruction.put((byte) (b >> 8));
    mmioInstruction.put((byte) b);


    return this;
  }

  public Trace mmioStamp(final long b) {
    if (filled.get(53)) {
      throw new IllegalStateException("mmio.MMIO_STAMP already set");
    } else {
      filled.set(53);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmio.MMIO_STAMP has invalid value (" + b + ")"); }
    mmioStamp.put((byte) (b >> 24));
    mmioStamp.put((byte) (b >> 16));
    mmioStamp.put((byte) (b >> 8));
    mmioStamp.put((byte) b);


    return this;
  }

  public Trace phase(final long b) {
    if (filled.get(54)) {
      throw new IllegalStateException("mmio.PHASE already set");
    } else {
      filled.set(54);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("mmio.PHASE has invalid value (" + b + ")"); }
    phase.put((byte) (b >> 24));
    phase.put((byte) (b >> 16));
    phase.put((byte) (b >> 8));
    phase.put((byte) b);


    return this;
  }

  public Trace pow2561(final Bytes b) {
    if (filled.get(55)) {
      throw new IllegalStateException("mmio.POW_256_1 already set");
    } else {
      filled.set(55);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.POW_256_1 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { pow2561.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { pow2561.put(bs.get(j)); }

    return this;
  }

  public Trace pow2562(final Bytes b) {
    if (filled.get(56)) {
      throw new IllegalStateException("mmio.POW_256_2 already set");
    } else {
      filled.set(56);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.POW_256_2 has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { pow2562.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { pow2562.put(bs.get(j)); }

    return this;
  }

  public Trace size(final Bytes b) {
    if (filled.get(57)) {
      throw new IllegalStateException("mmio.SIZE already set");
    } else {
      filled.set(57);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.SIZE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { size.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { size.put(bs.get(j)); }

    return this;
  }

  public Trace slow(final Boolean b) {
    if (filled.get(58)) {
      throw new IllegalStateException("mmio.SLOW already set");
    } else {
      filled.set(58);
    }

    slow.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace sourceByteOffset(final long b) {
    if (filled.get(59)) {
      throw new IllegalStateException("mmio.SOURCE_BYTE_OFFSET already set");
    } else {
      filled.set(59);
    }

    if(b >= 256L) { throw new IllegalArgumentException("mmio.SOURCE_BYTE_OFFSET has invalid value (" + b + ")"); }
    sourceByteOffset.put((byte) b);


    return this;
  }

  public Trace sourceLimbOffset(final Bytes b) {
    if (filled.get(60)) {
      throw new IllegalStateException("mmio.SOURCE_LIMB_OFFSET already set");
    } else {
      filled.set(60);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.SOURCE_LIMB_OFFSET has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { sourceLimbOffset.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { sourceLimbOffset.put(bs.get(j)); }

    return this;
  }

  public Trace successBit(final Boolean b) {
    if (filled.get(61)) {
      throw new IllegalStateException("mmio.SUCCESS_BIT already set");
    } else {
      filled.set(61);
    }

    successBit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace targetByteOffset(final long b) {
    if (filled.get(62)) {
      throw new IllegalStateException("mmio.TARGET_BYTE_OFFSET already set");
    } else {
      filled.set(62);
    }

    if(b >= 256L) { throw new IllegalArgumentException("mmio.TARGET_BYTE_OFFSET has invalid value (" + b + ")"); }
    targetByteOffset.put((byte) b);


    return this;
  }

  public Trace targetLimbOffset(final Bytes b) {
    if (filled.get(63)) {
      throw new IllegalStateException("mmio.TARGET_LIMB_OFFSET already set");
    } else {
      filled.set(63);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.TARGET_LIMB_OFFSET has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { targetLimbOffset.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { targetLimbOffset.put(bs.get(j)); }

    return this;
  }

  public Trace totalSize(final Bytes b) {
    if (filled.get(64)) {
      throw new IllegalStateException("mmio.TOTAL_SIZE already set");
    } else {
      filled.set(64);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 64) { throw new IllegalArgumentException("mmio.TOTAL_SIZE has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<8; i++) { totalSize.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { totalSize.put(bs.get(j)); }

    return this;
  }

  public Trace valA(final Bytes b) {
    if (filled.get(65)) {
      throw new IllegalStateException("mmio.VAL_A already set");
    } else {
      filled.set(65);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.VAL_A has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valA.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valA.put(bs.get(j)); }

    return this;
  }

  public Trace valANew(final Bytes b) {
    if (filled.get(66)) {
      throw new IllegalStateException("mmio.VAL_A_NEW already set");
    } else {
      filled.set(66);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.VAL_A_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valANew.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valANew.put(bs.get(j)); }

    return this;
  }

  public Trace valB(final Bytes b) {
    if (filled.get(67)) {
      throw new IllegalStateException("mmio.VAL_B already set");
    } else {
      filled.set(67);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.VAL_B has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valB.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valB.put(bs.get(j)); }

    return this;
  }

  public Trace valBNew(final Bytes b) {
    if (filled.get(68)) {
      throw new IllegalStateException("mmio.VAL_B_NEW already set");
    } else {
      filled.set(68);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.VAL_B_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valBNew.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valBNew.put(bs.get(j)); }

    return this;
  }

  public Trace valC(final Bytes b) {
    if (filled.get(69)) {
      throw new IllegalStateException("mmio.VAL_C already set");
    } else {
      filled.set(69);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.VAL_C has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valC.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valC.put(bs.get(j)); }

    return this;
  }

  public Trace valCNew(final Bytes b) {
    if (filled.get(70)) {
      throw new IllegalStateException("mmio.VAL_C_NEW already set");
    } else {
      filled.set(70);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("mmio.VAL_C_NEW has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { valCNew.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { valCNew.put(bs.get(j)); }

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("mmio.ACC_1 has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("mmio.ACC_2 has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("mmio.ACC_3 has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("mmio.ACC_4 has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("mmio.ACC_A has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("mmio.ACC_B has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("mmio.ACC_C has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("mmio.ACC_LIMB has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("mmio.BIT_1 has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("mmio.BIT_2 has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("mmio.BIT_3 has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("mmio.BIT_4 has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("mmio.BIT_5 has not been filled");
    }

    if (!filled.get(13)) {
      throw new IllegalStateException("mmio.BYTE_A has not been filled");
    }

    if (!filled.get(14)) {
      throw new IllegalStateException("mmio.BYTE_B has not been filled");
    }

    if (!filled.get(15)) {
      throw new IllegalStateException("mmio.BYTE_C has not been filled");
    }

    if (!filled.get(16)) {
      throw new IllegalStateException("mmio.BYTE_LIMB has not been filled");
    }

    if (!filled.get(17)) {
      throw new IllegalStateException("mmio.CN_A has not been filled");
    }

    if (!filled.get(18)) {
      throw new IllegalStateException("mmio.CN_B has not been filled");
    }

    if (!filled.get(19)) {
      throw new IllegalStateException("mmio.CN_C has not been filled");
    }

    if (!filled.get(20)) {
      throw new IllegalStateException("mmio.CONTEXT_SOURCE has not been filled");
    }

    if (!filled.get(21)) {
      throw new IllegalStateException("mmio.CONTEXT_TARGET has not been filled");
    }

    if (!filled.get(22)) {
      throw new IllegalStateException("mmio.COUNTER has not been filled");
    }

    if (!filled.get(23)) {
      throw new IllegalStateException("mmio.EXO_ID has not been filled");
    }

    if (!filled.get(24)) {
      throw new IllegalStateException("mmio.EXO_IS_BLAKEMODEXP has not been filled");
    }

    if (!filled.get(25)) {
      throw new IllegalStateException("mmio.EXO_IS_ECDATA has not been filled");
    }

    if (!filled.get(26)) {
      throw new IllegalStateException("mmio.EXO_IS_KEC has not been filled");
    }

    if (!filled.get(27)) {
      throw new IllegalStateException("mmio.EXO_IS_LOG has not been filled");
    }

    if (!filled.get(28)) {
      throw new IllegalStateException("mmio.EXO_IS_RIPSHA has not been filled");
    }

    if (!filled.get(29)) {
      throw new IllegalStateException("mmio.EXO_IS_ROM has not been filled");
    }

    if (!filled.get(30)) {
      throw new IllegalStateException("mmio.EXO_IS_TXCD has not been filled");
    }

    if (!filled.get(31)) {
      throw new IllegalStateException("mmio.EXO_SUM has not been filled");
    }

    if (!filled.get(32)) {
      throw new IllegalStateException("mmio.FAST has not been filled");
    }

    if (!filled.get(33)) {
      throw new IllegalStateException("mmio.INDEX_A has not been filled");
    }

    if (!filled.get(34)) {
      throw new IllegalStateException("mmio.INDEX_B has not been filled");
    }

    if (!filled.get(35)) {
      throw new IllegalStateException("mmio.INDEX_C has not been filled");
    }

    if (!filled.get(36)) {
      throw new IllegalStateException("mmio.INDEX_X has not been filled");
    }

    if (!filled.get(37)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_ONE_TARGET has not been filled");
    }

    if (!filled.get(38)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TRANSPLANT has not been filled");
    }

    if (!filled.get(39)) {
      throw new IllegalStateException("mmio.IS_LIMB_TO_RAM_TWO_TARGET has not been filled");
    }

    if (!filled.get(40)) {
      throw new IllegalStateException("mmio.IS_LIMB_VANISHES has not been filled");
    }

    if (!filled.get(41)) {
      throw new IllegalStateException("mmio.IS_RAM_EXCISION has not been filled");
    }

    if (!filled.get(42)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_ONE_SOURCE has not been filled");
    }

    if (!filled.get(43)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TRANSPLANT has not been filled");
    }

    if (!filled.get(44)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_LIMB_TWO_SOURCE has not been filled");
    }

    if (!filled.get(45)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_PARTIAL has not been filled");
    }

    if (!filled.get(46)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TRANSPLANT has not been filled");
    }

    if (!filled.get(47)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_SOURCE has not been filled");
    }

    if (!filled.get(48)) {
      throw new IllegalStateException("mmio.IS_RAM_TO_RAM_TWO_TARGET has not been filled");
    }

    if (!filled.get(49)) {
      throw new IllegalStateException("mmio.IS_RAM_VANISHES has not been filled");
    }

    if (!filled.get(50)) {
      throw new IllegalStateException("mmio.KEC_ID has not been filled");
    }

    if (!filled.get(51)) {
      throw new IllegalStateException("mmio.LIMB has not been filled");
    }

    if (!filled.get(52)) {
      throw new IllegalStateException("mmio.MMIO_INSTRUCTION has not been filled");
    }

    if (!filled.get(53)) {
      throw new IllegalStateException("mmio.MMIO_STAMP has not been filled");
    }

    if (!filled.get(54)) {
      throw new IllegalStateException("mmio.PHASE has not been filled");
    }

    if (!filled.get(55)) {
      throw new IllegalStateException("mmio.POW_256_1 has not been filled");
    }

    if (!filled.get(56)) {
      throw new IllegalStateException("mmio.POW_256_2 has not been filled");
    }

    if (!filled.get(57)) {
      throw new IllegalStateException("mmio.SIZE has not been filled");
    }

    if (!filled.get(58)) {
      throw new IllegalStateException("mmio.SLOW has not been filled");
    }

    if (!filled.get(59)) {
      throw new IllegalStateException("mmio.SOURCE_BYTE_OFFSET has not been filled");
    }

    if (!filled.get(60)) {
      throw new IllegalStateException("mmio.SOURCE_LIMB_OFFSET has not been filled");
    }

    if (!filled.get(61)) {
      throw new IllegalStateException("mmio.SUCCESS_BIT has not been filled");
    }

    if (!filled.get(62)) {
      throw new IllegalStateException("mmio.TARGET_BYTE_OFFSET has not been filled");
    }

    if (!filled.get(63)) {
      throw new IllegalStateException("mmio.TARGET_LIMB_OFFSET has not been filled");
    }

    if (!filled.get(64)) {
      throw new IllegalStateException("mmio.TOTAL_SIZE has not been filled");
    }

    if (!filled.get(65)) {
      throw new IllegalStateException("mmio.VAL_A has not been filled");
    }

    if (!filled.get(66)) {
      throw new IllegalStateException("mmio.VAL_A_NEW has not been filled");
    }

    if (!filled.get(67)) {
      throw new IllegalStateException("mmio.VAL_B has not been filled");
    }

    if (!filled.get(68)) {
      throw new IllegalStateException("mmio.VAL_B_NEW has not been filled");
    }

    if (!filled.get(69)) {
      throw new IllegalStateException("mmio.VAL_C has not been filled");
    }

    if (!filled.get(70)) {
      throw new IllegalStateException("mmio.VAL_C_NEW has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      acc1.position(acc1.position() + 16);
    }

    if (!filled.get(1)) {
      acc2.position(acc2.position() + 16);
    }

    if (!filled.get(2)) {
      acc3.position(acc3.position() + 16);
    }

    if (!filled.get(3)) {
      acc4.position(acc4.position() + 16);
    }

    if (!filled.get(4)) {
      accA.position(accA.position() + 16);
    }

    if (!filled.get(5)) {
      accB.position(accB.position() + 16);
    }

    if (!filled.get(6)) {
      accC.position(accC.position() + 16);
    }

    if (!filled.get(7)) {
      accLimb.position(accLimb.position() + 16);
    }

    if (!filled.get(8)) {
      bit1.position(bit1.position() + 1);
    }

    if (!filled.get(9)) {
      bit2.position(bit2.position() + 1);
    }

    if (!filled.get(10)) {
      bit3.position(bit3.position() + 1);
    }

    if (!filled.get(11)) {
      bit4.position(bit4.position() + 1);
    }

    if (!filled.get(12)) {
      bit5.position(bit5.position() + 1);
    }

    if (!filled.get(13)) {
      byteA.position(byteA.position() + 1);
    }

    if (!filled.get(14)) {
      byteB.position(byteB.position() + 1);
    }

    if (!filled.get(15)) {
      byteC.position(byteC.position() + 1);
    }

    if (!filled.get(16)) {
      byteLimb.position(byteLimb.position() + 1);
    }

    if (!filled.get(17)) {
      cnA.position(cnA.position() + 8);
    }

    if (!filled.get(18)) {
      cnB.position(cnB.position() + 8);
    }

    if (!filled.get(19)) {
      cnC.position(cnC.position() + 8);
    }

    if (!filled.get(20)) {
      contextSource.position(contextSource.position() + 8);
    }

    if (!filled.get(21)) {
      contextTarget.position(contextTarget.position() + 8);
    }

    if (!filled.get(22)) {
      counter.position(counter.position() + 1);
    }

    if (!filled.get(23)) {
      exoId.position(exoId.position() + 4);
    }

    if (!filled.get(24)) {
      exoIsBlakemodexp.position(exoIsBlakemodexp.position() + 1);
    }

    if (!filled.get(25)) {
      exoIsEcdata.position(exoIsEcdata.position() + 1);
    }

    if (!filled.get(26)) {
      exoIsKec.position(exoIsKec.position() + 1);
    }

    if (!filled.get(27)) {
      exoIsLog.position(exoIsLog.position() + 1);
    }

    if (!filled.get(28)) {
      exoIsRipsha.position(exoIsRipsha.position() + 1);
    }

    if (!filled.get(29)) {
      exoIsRom.position(exoIsRom.position() + 1);
    }

    if (!filled.get(30)) {
      exoIsTxcd.position(exoIsTxcd.position() + 1);
    }

    if (!filled.get(31)) {
      exoSum.position(exoSum.position() + 4);
    }

    if (!filled.get(32)) {
      fast.position(fast.position() + 1);
    }

    if (!filled.get(33)) {
      indexA.position(indexA.position() + 8);
    }

    if (!filled.get(34)) {
      indexB.position(indexB.position() + 8);
    }

    if (!filled.get(35)) {
      indexC.position(indexC.position() + 8);
    }

    if (!filled.get(36)) {
      indexX.position(indexX.position() + 8);
    }

    if (!filled.get(37)) {
      isLimbToRamOneTarget.position(isLimbToRamOneTarget.position() + 1);
    }

    if (!filled.get(38)) {
      isLimbToRamTransplant.position(isLimbToRamTransplant.position() + 1);
    }

    if (!filled.get(39)) {
      isLimbToRamTwoTarget.position(isLimbToRamTwoTarget.position() + 1);
    }

    if (!filled.get(40)) {
      isLimbVanishes.position(isLimbVanishes.position() + 1);
    }

    if (!filled.get(41)) {
      isRamExcision.position(isRamExcision.position() + 1);
    }

    if (!filled.get(42)) {
      isRamToLimbOneSource.position(isRamToLimbOneSource.position() + 1);
    }

    if (!filled.get(43)) {
      isRamToLimbTransplant.position(isRamToLimbTransplant.position() + 1);
    }

    if (!filled.get(44)) {
      isRamToLimbTwoSource.position(isRamToLimbTwoSource.position() + 1);
    }

    if (!filled.get(45)) {
      isRamToRamPartial.position(isRamToRamPartial.position() + 1);
    }

    if (!filled.get(46)) {
      isRamToRamTransplant.position(isRamToRamTransplant.position() + 1);
    }

    if (!filled.get(47)) {
      isRamToRamTwoSource.position(isRamToRamTwoSource.position() + 1);
    }

    if (!filled.get(48)) {
      isRamToRamTwoTarget.position(isRamToRamTwoTarget.position() + 1);
    }

    if (!filled.get(49)) {
      isRamVanishes.position(isRamVanishes.position() + 1);
    }

    if (!filled.get(50)) {
      kecId.position(kecId.position() + 4);
    }

    if (!filled.get(51)) {
      limb.position(limb.position() + 16);
    }

    if (!filled.get(52)) {
      mmioInstruction.position(mmioInstruction.position() + 2);
    }

    if (!filled.get(53)) {
      mmioStamp.position(mmioStamp.position() + 4);
    }

    if (!filled.get(54)) {
      phase.position(phase.position() + 4);
    }

    if (!filled.get(55)) {
      pow2561.position(pow2561.position() + 16);
    }

    if (!filled.get(56)) {
      pow2562.position(pow2562.position() + 16);
    }

    if (!filled.get(57)) {
      size.position(size.position() + 8);
    }

    if (!filled.get(58)) {
      slow.position(slow.position() + 1);
    }

    if (!filled.get(59)) {
      sourceByteOffset.position(sourceByteOffset.position() + 1);
    }

    if (!filled.get(60)) {
      sourceLimbOffset.position(sourceLimbOffset.position() + 8);
    }

    if (!filled.get(61)) {
      successBit.position(successBit.position() + 1);
    }

    if (!filled.get(62)) {
      targetByteOffset.position(targetByteOffset.position() + 1);
    }

    if (!filled.get(63)) {
      targetLimbOffset.position(targetLimbOffset.position() + 8);
    }

    if (!filled.get(64)) {
      totalSize.position(totalSize.position() + 8);
    }

    if (!filled.get(65)) {
      valA.position(valA.position() + 16);
    }

    if (!filled.get(66)) {
      valANew.position(valANew.position() + 16);
    }

    if (!filled.get(67)) {
      valB.position(valB.position() + 16);
    }

    if (!filled.get(68)) {
      valBNew.position(valBNew.position() + 16);
    }

    if (!filled.get(69)) {
      valC.position(valC.position() + 16);
    }

    if (!filled.get(70)) {
      valCNew.position(valCNew.position() + 16);
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
