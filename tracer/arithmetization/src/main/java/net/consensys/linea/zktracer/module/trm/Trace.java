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

package net.consensys.linea.zktracer.module.trm;

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

  private final MappedByteBuffer accHi;
  private final MappedByteBuffer accLo;
  private final MappedByteBuffer accT;
  private final MappedByteBuffer addrHi;
  private final MappedByteBuffer addrLo;
  private final MappedByteBuffer byteHi;
  private final MappedByteBuffer byteLo;
  private final MappedByteBuffer ct;
  private final MappedByteBuffer isPrec;
  private final MappedByteBuffer one;
  private final MappedByteBuffer pbit;
  private final MappedByteBuffer stamp;
  private final MappedByteBuffer trmAddrHi;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("trm.ACC_HI", 32, length),
        new ColumnHeader("trm.ACC_LO", 32, length),
        new ColumnHeader("trm.ACC_T", 32, length),
        new ColumnHeader("trm.ADDR_HI", 32, length),
        new ColumnHeader("trm.ADDR_LO", 32, length),
        new ColumnHeader("trm.BYTE_HI", 1, length),
        new ColumnHeader("trm.BYTE_LO", 1, length),
        new ColumnHeader("trm.CT", 32, length),
        new ColumnHeader("trm.IS_PREC", 1, length),
        new ColumnHeader("trm.ONE", 1, length),
        new ColumnHeader("trm.PBIT", 1, length),
        new ColumnHeader("trm.STAMP", 32, length),
        new ColumnHeader("trm.TRM_ADDR_HI", 32, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.accHi = buffers.get(0);
    this.accLo = buffers.get(1);
    this.accT = buffers.get(2);
    this.addrHi = buffers.get(3);
    this.addrLo = buffers.get(4);
    this.byteHi = buffers.get(5);
    this.byteLo = buffers.get(6);
    this.ct = buffers.get(7);
    this.isPrec = buffers.get(8);
    this.one = buffers.get(9);
    this.pbit = buffers.get(10);
    this.stamp = buffers.get(11);
    this.trmAddrHi = buffers.get(12);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace accHi(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("trm.ACC_HI already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accHi.put((byte) 0);
    }
    accHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accLo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("trm.ACC_LO already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accLo.put((byte) 0);
    }
    accLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace accT(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("trm.ACC_T already set");
    } else {
      filled.set(2);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      accT.put((byte) 0);
    }
    accT.put(b.toArrayUnsafe());

    return this;
  }

  public Trace addrHi(final Bytes b) {
    if (filled.get(3)) {
      throw new IllegalStateException("trm.ADDR_HI already set");
    } else {
      filled.set(3);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrHi.put((byte) 0);
    }
    addrHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("trm.ADDR_LO already set");
    } else {
      filled.set(4);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrLo.put((byte) 0);
    }
    addrLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace byteHi(final UnsignedByte b) {
    if (filled.get(5)) {
      throw new IllegalStateException("trm.BYTE_HI already set");
    } else {
      filled.set(5);
    }

    byteHi.put(b.toByte());

    return this;
  }

  public Trace byteLo(final UnsignedByte b) {
    if (filled.get(6)) {
      throw new IllegalStateException("trm.BYTE_LO already set");
    } else {
      filled.set(6);
    }

    byteLo.put(b.toByte());

    return this;
  }

  public Trace ct(final Bytes b) {
    if (filled.get(7)) {
      throw new IllegalStateException("trm.CT already set");
    } else {
      filled.set(7);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      ct.put((byte) 0);
    }
    ct.put(b.toArrayUnsafe());

    return this;
  }

  public Trace isPrec(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("trm.IS_PREC already set");
    } else {
      filled.set(8);
    }

    isPrec.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace one(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("trm.ONE already set");
    } else {
      filled.set(9);
    }

    one.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace pbit(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("trm.PBIT already set");
    } else {
      filled.set(10);
    }

    pbit.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace stamp(final Bytes b) {
    if (filled.get(11)) {
      throw new IllegalStateException("trm.STAMP already set");
    } else {
      filled.set(11);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      stamp.put((byte) 0);
    }
    stamp.put(b.toArrayUnsafe());

    return this;
  }

  public Trace trmAddrHi(final Bytes b) {
    if (filled.get(12)) {
      throw new IllegalStateException("trm.TRM_ADDR_HI already set");
    } else {
      filled.set(12);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      trmAddrHi.put((byte) 0);
    }
    trmAddrHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("trm.ACC_HI has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("trm.ACC_LO has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("trm.ACC_T has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("trm.ADDR_HI has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("trm.ADDR_LO has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("trm.BYTE_HI has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("trm.BYTE_LO has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("trm.CT has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("trm.IS_PREC has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("trm.ONE has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("trm.PBIT has not been filled");
    }

    if (!filled.get(11)) {
      throw new IllegalStateException("trm.STAMP has not been filled");
    }

    if (!filled.get(12)) {
      throw new IllegalStateException("trm.TRM_ADDR_HI has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      accHi.position(accHi.position() + 32);
    }

    if (!filled.get(1)) {
      accLo.position(accLo.position() + 32);
    }

    if (!filled.get(2)) {
      accT.position(accT.position() + 32);
    }

    if (!filled.get(3)) {
      addrHi.position(addrHi.position() + 32);
    }

    if (!filled.get(4)) {
      addrLo.position(addrLo.position() + 32);
    }

    if (!filled.get(5)) {
      byteHi.position(byteHi.position() + 1);
    }

    if (!filled.get(6)) {
      byteLo.position(byteLo.position() + 1);
    }

    if (!filled.get(7)) {
      ct.position(ct.position() + 32);
    }

    if (!filled.get(8)) {
      isPrec.position(isPrec.position() + 1);
    }

    if (!filled.get(9)) {
      one.position(one.position() + 1);
    }

    if (!filled.get(10)) {
      pbit.position(pbit.position() + 1);
    }

    if (!filled.get(11)) {
      stamp.position(stamp.position() + 32);
    }

    if (!filled.get(12)) {
      trmAddrHi.position(trmAddrHi.position() + 32);
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
