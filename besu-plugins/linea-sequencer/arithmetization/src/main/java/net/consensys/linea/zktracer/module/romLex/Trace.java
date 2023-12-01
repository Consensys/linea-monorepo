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

package net.consensys.linea.zktracer.module.romLex;

import java.nio.MappedByteBuffer;
import java.util.BitSet;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
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

  private final MappedByteBuffer addrHi;
  private final MappedByteBuffer addrLo;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexInfty;
  private final MappedByteBuffer codeSize;
  private final MappedByteBuffer commitToState;
  private final MappedByteBuffer depNumber;
  private final MappedByteBuffer depStatus;
  private final MappedByteBuffer readFromState;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("romLex.ADDR_HI", 32, length),
        new ColumnHeader("romLex.ADDR_LO", 32, length),
        new ColumnHeader("romLex.CODE_FRAGMENT_INDEX", 32, length),
        new ColumnHeader("romLex.CODE_FRAGMENT_INDEX_INFTY", 32, length),
        new ColumnHeader("romLex.CODE_SIZE", 32, length),
        new ColumnHeader("romLex.COMMIT_TO_STATE", 1, length),
        new ColumnHeader("romLex.DEP_NUMBER", 32, length),
        new ColumnHeader("romLex.DEP_STATUS", 1, length),
        new ColumnHeader("romLex.READ_FROM_STATE", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.addrHi = buffers.get(0);
    this.addrLo = buffers.get(1);
    this.codeFragmentIndex = buffers.get(2);
    this.codeFragmentIndexInfty = buffers.get(3);
    this.codeSize = buffers.get(4);
    this.commitToState = buffers.get(5);
    this.depNumber = buffers.get(6);
    this.depStatus = buffers.get(7);
    this.readFromState = buffers.get(8);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace addrHi(final Bytes b) {
    if (filled.get(0)) {
      throw new IllegalStateException("romLex.ADDR_HI already set");
    } else {
      filled.set(0);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrHi.put((byte) 0);
    }
    addrHi.put(b.toArrayUnsafe());

    return this;
  }

  public Trace addrLo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("romLex.ADDR_LO already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addrLo.put((byte) 0);
    }
    addrLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndex(final Bytes b) {
    if (filled.get(2)) {
      throw new IllegalStateException("romLex.CODE_FRAGMENT_INDEX already set");
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
      throw new IllegalStateException("romLex.CODE_FRAGMENT_INDEX_INFTY already set");
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
      throw new IllegalStateException("romLex.CODE_SIZE already set");
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

  public Trace commitToState(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("romLex.COMMIT_TO_STATE already set");
    } else {
      filled.set(5);
    }

    commitToState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace depNumber(final Bytes b) {
    if (filled.get(6)) {
      throw new IllegalStateException("romLex.DEP_NUMBER already set");
    } else {
      filled.set(6);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      depNumber.put((byte) 0);
    }
    depNumber.put(b.toArrayUnsafe());

    return this;
  }

  public Trace depStatus(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("romLex.DEP_STATUS already set");
    } else {
      filled.set(7);
    }

    depStatus.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace readFromState(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("romLex.READ_FROM_STATE already set");
    } else {
      filled.set(8);
    }

    readFromState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("romLex.ADDR_HI has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("romLex.ADDR_LO has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("romLex.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("romLex.CODE_FRAGMENT_INDEX_INFTY has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("romLex.CODE_SIZE has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("romLex.COMMIT_TO_STATE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("romLex.DEP_NUMBER has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("romLex.DEP_STATUS has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("romLex.READ_FROM_STATE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      addrHi.position(addrHi.position() + 32);
    }

    if (!filled.get(1)) {
      addrLo.position(addrLo.position() + 32);
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

    if (!filled.get(5)) {
      commitToState.position(commitToState.position() + 1);
    }

    if (!filled.get(6)) {
      depNumber.position(depNumber.position() + 32);
    }

    if (!filled.get(7)) {
      depStatus.position(depStatus.position() + 1);
    }

    if (!filled.get(8)) {
      readFromState.position(readFromState.position() + 1);
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
