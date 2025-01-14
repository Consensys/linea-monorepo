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

  private final MappedByteBuffer addressHi;
  private final MappedByteBuffer addressLo;
  private final MappedByteBuffer codeFragmentIndex;
  private final MappedByteBuffer codeFragmentIndexInfty;
  private final MappedByteBuffer codeHashHi;
  private final MappedByteBuffer codeHashLo;
  private final MappedByteBuffer codeSize;
  private final MappedByteBuffer commitToState;
  private final MappedByteBuffer deploymentNumber;
  private final MappedByteBuffer deploymentStatus;
  private final MappedByteBuffer readFromState;

  static List<ColumnHeader> headers(int length) {
      List<ColumnHeader> headers = new ArrayList<>();
      headers.add(new ColumnHeader("romlex.ADDRESS_HI", 4, length));
      headers.add(new ColumnHeader("romlex.ADDRESS_LO", 16, length));
      headers.add(new ColumnHeader("romlex.CODE_FRAGMENT_INDEX", 4, length));
      headers.add(new ColumnHeader("romlex.CODE_FRAGMENT_INDEX_INFTY", 4, length));
      headers.add(new ColumnHeader("romlex.CODE_HASH_HI", 16, length));
      headers.add(new ColumnHeader("romlex.CODE_HASH_LO", 16, length));
      headers.add(new ColumnHeader("romlex.CODE_SIZE", 4, length));
      headers.add(new ColumnHeader("romlex.COMMIT_TO_STATE", 1, length));
      headers.add(new ColumnHeader("romlex.DEPLOYMENT_NUMBER", 2, length));
      headers.add(new ColumnHeader("romlex.DEPLOYMENT_STATUS", 1, length));
      headers.add(new ColumnHeader("romlex.READ_FROM_STATE", 1, length));
      return headers;
  }

  public Trace (List<MappedByteBuffer> buffers) {
    this.addressHi = buffers.get(0);
    this.addressLo = buffers.get(1);
    this.codeFragmentIndex = buffers.get(2);
    this.codeFragmentIndexInfty = buffers.get(3);
    this.codeHashHi = buffers.get(4);
    this.codeHashLo = buffers.get(5);
    this.codeSize = buffers.get(6);
    this.commitToState = buffers.get(7);
    this.deploymentNumber = buffers.get(8);
    this.deploymentStatus = buffers.get(9);
    this.readFromState = buffers.get(10);
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

    if(b >= 4294967296L) { throw new IllegalArgumentException("romlex.ADDRESS_HI has invalid value (" + b + ")"); }
    addressHi.put((byte) (b >> 24));
    addressHi.put((byte) (b >> 16));
    addressHi.put((byte) (b >> 8));
    addressHi.put((byte) b);


    return this;
  }

  public Trace addressLo(final Bytes b) {
    if (filled.get(1)) {
      throw new IllegalStateException("romlex.ADDRESS_LO already set");
    } else {
      filled.set(1);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("romlex.ADDRESS_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { addressLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { addressLo.put(bs.get(j)); }

    return this;
  }

  public Trace codeFragmentIndex(final long b) {
    if (filled.get(2)) {
      throw new IllegalStateException("romlex.CODE_FRAGMENT_INDEX already set");
    } else {
      filled.set(2);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("romlex.CODE_FRAGMENT_INDEX has invalid value (" + b + ")"); }
    codeFragmentIndex.put((byte) (b >> 24));
    codeFragmentIndex.put((byte) (b >> 16));
    codeFragmentIndex.put((byte) (b >> 8));
    codeFragmentIndex.put((byte) b);


    return this;
  }

  public Trace codeFragmentIndexInfty(final long b) {
    if (filled.get(3)) {
      throw new IllegalStateException("romlex.CODE_FRAGMENT_INDEX_INFTY already set");
    } else {
      filled.set(3);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("romlex.CODE_FRAGMENT_INDEX_INFTY has invalid value (" + b + ")"); }
    codeFragmentIndexInfty.put((byte) (b >> 24));
    codeFragmentIndexInfty.put((byte) (b >> 16));
    codeFragmentIndexInfty.put((byte) (b >> 8));
    codeFragmentIndexInfty.put((byte) b);


    return this;
  }

  public Trace codeHashHi(final Bytes b) {
    if (filled.get(4)) {
      throw new IllegalStateException("romlex.CODE_HASH_HI already set");
    } else {
      filled.set(4);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("romlex.CODE_HASH_HI has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashHi.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashHi.put(bs.get(j)); }

    return this;
  }

  public Trace codeHashLo(final Bytes b) {
    if (filled.get(5)) {
      throw new IllegalStateException("romlex.CODE_HASH_LO already set");
    } else {
      filled.set(5);
    }

    // Trim array to size
    Bytes bs = b.trimLeadingZeros();
    // Sanity check against expected width
    if(bs.bitLength() > 128) { throw new IllegalArgumentException("romlex.CODE_HASH_LO has invalid width (" + bs.bitLength() + "bits)"); }
    // Write padding (if necessary)
    for(int i=bs.size(); i<16; i++) { codeHashLo.put((byte) 0); }
    // Write bytes
    for(int j=0; j<bs.size(); j++) { codeHashLo.put(bs.get(j)); }

    return this;
  }

  public Trace codeSize(final long b) {
    if (filled.get(6)) {
      throw new IllegalStateException("romlex.CODE_SIZE already set");
    } else {
      filled.set(6);
    }

    if(b >= 4294967296L) { throw new IllegalArgumentException("romlex.CODE_SIZE has invalid value (" + b + ")"); }
    codeSize.put((byte) (b >> 24));
    codeSize.put((byte) (b >> 16));
    codeSize.put((byte) (b >> 8));
    codeSize.put((byte) b);


    return this;
  }

  public Trace commitToState(final Boolean b) {
    if (filled.get(7)) {
      throw new IllegalStateException("romlex.COMMIT_TO_STATE already set");
    } else {
      filled.set(7);
    }

    commitToState.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace deploymentNumber(final long b) {
    if (filled.get(8)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_NUMBER already set");
    } else {
      filled.set(8);
    }

    if(b >= 65536L) { throw new IllegalArgumentException("romlex.DEPLOYMENT_NUMBER has invalid value (" + b + ")"); }
    deploymentNumber.put((byte) (b >> 8));
    deploymentNumber.put((byte) b);


    return this;
  }

  public Trace deploymentStatus(final Boolean b) {
    if (filled.get(9)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_STATUS already set");
    } else {
      filled.set(9);
    }

    deploymentStatus.put((byte) (b ? 1 : 0));

    return this;
  }

  public Trace readFromState(final Boolean b) {
    if (filled.get(10)) {
      throw new IllegalStateException("romlex.READ_FROM_STATE already set");
    } else {
      filled.set(10);
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
      throw new IllegalStateException("romlex.CODE_HASH_HI has not been filled");
    }

    if (!filled.get(5)) {
      throw new IllegalStateException("romlex.CODE_HASH_LO has not been filled");
    }

    if (!filled.get(6)) {
      throw new IllegalStateException("romlex.CODE_SIZE has not been filled");
    }

    if (!filled.get(7)) {
      throw new IllegalStateException("romlex.COMMIT_TO_STATE has not been filled");
    }

    if (!filled.get(8)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_NUMBER has not been filled");
    }

    if (!filled.get(9)) {
      throw new IllegalStateException("romlex.DEPLOYMENT_STATUS has not been filled");
    }

    if (!filled.get(10)) {
      throw new IllegalStateException("romlex.READ_FROM_STATE has not been filled");
    }

    filled.clear();
    this.currentLine++;

    return this;
  }

  public Trace fillAndValidateRow() {
    if (!filled.get(0)) {
      addressHi.position(addressHi.position() + 4);
    }

    if (!filled.get(1)) {
      addressLo.position(addressLo.position() + 16);
    }

    if (!filled.get(2)) {
      codeFragmentIndex.position(codeFragmentIndex.position() + 4);
    }

    if (!filled.get(3)) {
      codeFragmentIndexInfty.position(codeFragmentIndexInfty.position() + 4);
    }

    if (!filled.get(4)) {
      codeHashHi.position(codeHashHi.position() + 16);
    }

    if (!filled.get(5)) {
      codeHashLo.position(codeHashLo.position() + 16);
    }

    if (!filled.get(6)) {
      codeSize.position(codeSize.position() + 4);
    }

    if (!filled.get(7)) {
      commitToState.position(commitToState.position() + 1);
    }

    if (!filled.get(8)) {
      deploymentNumber.position(deploymentNumber.position() + 2);
    }

    if (!filled.get(9)) {
      deploymentStatus.position(deploymentStatus.position() + 1);
    }

    if (!filled.get(10)) {
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
