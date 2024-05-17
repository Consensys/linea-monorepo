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

package net.consensys.linea.zktracer.module.romlex;

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

  private final MappedByteBuffer addressHi;
  private final MappedByteBuffer addressLo;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexInfty;
  private final MappedByteBuffer codeSize;
  private final MappedByteBuffer commitToState;
  private final MappedByteBuffer deploymentNumber;
  private final MappedByteBuffer deploymentStatus;
  private final MappedByteBuffer readFromState;

  static List<ColumnHeader> headers(int length) {
    return List.of(
        new ColumnHeader("romlex.ADDRESS_HI", 8, length),
        new ColumnHeader("romlex.ADDRESS_LO", 32, length),
        new ColumnHeader("romlex.CODE_FRAGMENT_INDEX", 8, length),
        new ColumnHeader("romlex.CODE_FRAGMENT_INDEX_INFTY", 8, length),
        new ColumnHeader("romlex.CODE_SIZE", 8, length),
        new ColumnHeader("romlex.COMMIT_TO_STATE", 1, length),
        new ColumnHeader("romlex.DEPLOYMENT_NUMBER", 4, length),
        new ColumnHeader("romlex.DEPLOYMENT_STATUS", 1, length),
        new ColumnHeader("romlex.READ_FROM_STATE", 1, length));
  }

  public Trace(List<MappedByteBuffer> buffers) {
    this.addressHi = buffers.get(0);
    this.addressLo = buffers.get(1);
    this.codeFragmentIndex = buffers.get(2);
    this.codeFragmentIndexInfty = buffers.get(3);
    this.codeSize = buffers.get(4);
    this.commitToState = buffers.get(5);
    this.deploymentNumber = buffers.get(6);
    this.deploymentStatus = buffers.get(7);
    this.readFromState = buffers.get(8);
  }

  public int size() {
    if (!filled.isEmpty()) {
      throw new RuntimeException("Cannot measure a trace with a non-validated row.");
    }

    return this.currentLine;
  }

  public Trace addressHi(final long b) {
    if (filled.get(0)) {
      throw new IllegalStateException("romlex.ADDRESS_HI already set");
    } else {
      filled.set(0);
    }

    addressHi.putLong(b);

    return this;
  }

  public Trace addressLo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("romlex.ADDRESS_LO already set");
    } else {
      filled.set(1);
    }

    final byte[] bs = b.toArrayUnsafe();
    for (int i = bs.length; i < 32; i++) {
      addressLo.put((byte) 0);
    }
    addressLo.put(b.toArrayUnsafe());

    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("romlex.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(2);
    }

    codeFragmentIndex.putLong(b);

    return this;
  }

  public Trace codeFragmentIndexInfty(final long b) {
    if (filled.get(3)) {
      throw new IllegalStateException("romlex.CODE_FRAGMENT_INDEX_INFTY already set");
    } else {
      filled.set(3);
    }

    codeFragmentIndexInfty.putLong(b);

    return this;
  }

  public Trace codeSize(final long b) {
    if (filled.get(4)) {
      throw new IllegalStateException("romlex.CODE_SIZE already set");
    } else {
      filled.set(4);
    }

    codeSize.putLong(b);

    return this;
  }

  public Trace commitToState(final Boolean b) {
    if (filled.get(5)) {
      throw new IllegalStateException("romlex.COMMIT_TO_STATE already set");
    } else {
      filled.set(5);
    }

    commitToState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace deploymentNumber(final int b) {
    if (filled.get(6)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(6);
    }

    deploymentNumber.putInt(b);

    return this;
  }

  public Trace deploymentStatus(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_STATUS already set");
    } else {
      filled.set(7);
    }

    deploymentStatus.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace readFromState(final Boolean b) {
    if (filled.get(8)) {
      throw new IllegalStateException("romlex.READ_FROM_STATE already set");
    } else {
      filled.set(8);
    }

    readFromState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace validateRow() {
    if (!filled.get(0)) {
      throw new IllegalStateException("romlex.ADDRESS_HI has not been filled");
    }

    if (!filled.get(1)) {
      throw new IllegalStateException("romlex.ADDRESS_LO has not been filled");
    }

    if (!filled.get(2)) {
      throw new IllegalStateException("romlex.CODE_FRAGMENT_INDEX has not been filled");
    }

    if (!filled.get(3)) {
      throw new IllegalStateException("romlex.CODE_FRAGMENT_INDEX_INFTY has not been filled");
    }

    if (!filled.get(4)) {
      throw new IllegalStateException("romlex.CODE_SIZE has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("romlex.COMMIT_TO_STATE has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_NUMBER has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_STATUS has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("romlex.READ_FROM_STATE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      addressHi.position(addressHi.position() + 8);
    }

    if (!filled.get(1)) {
      addressLo.position(addressLo.position() + 32);
    }

    if (!filled.get(2)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 8);
    }

    if (!filled.get(3)) {
      codeFragmentIndexInfty.position(codeFragmentIndexInfty.position() + 8);
    }

    if (!filled.get(4)) {
      codeSize.position(codeSize.position() + 8);
    }

    if (!filled.get(5)) {
      commitToState.position(commitToState.position() + 1);
    }

    if (!filled.get(6)) {
      deploymentNumber.position(deploymentNumber.position() + 4);
    }

    if (!filled.get(7)) {
      deploymentStatus.position(deploymentStatus.position() + 1);
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
