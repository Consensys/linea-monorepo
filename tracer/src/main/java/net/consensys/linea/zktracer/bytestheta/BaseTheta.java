package net.consensys.linea.zktracer.bytestheta;
/*
 * Copyright ConsenSys AG.
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

import net.consensys.linea.zktracer.bytes.Bytes16;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class BaseTheta extends BaseBytes {

  private BaseTheta(final Bytes32 arg) {
    super(arg);
    bytes32 = arg.mutableCopy();
    for (int k = 0; k < 4; k++) {
      Bytes bytes = arg.slice(OFFSET * k, OFFSET);
      setBytes(OFFSET * (3 - k), bytes);
    }
  }

  public static BaseTheta fromBytes32(Bytes32 arg) {
    return new BaseTheta(arg);
  }

  public void setBytes(int index, Bytes bytes) {
    bytes32.set(index, bytes);
  }

  // set the whole chunk of bytes at the given index.
  // assumes index is one of 0,1,2,3
  // assumes length of bytes is 8
  public void setChunk(int index, Bytes bytes) {
    bytes32.set(OFFSET * index, bytes);
  }

  public Bytes get(int index) {
    return bytes32.slice(OFFSET * index, OFFSET);
  }

  public Bytes slice(int i, int length) {
    return bytes32.slice(i, length);
  }

  @Override
  public Bytes16 getHigh() {
    return Bytes16.wrap(Bytes.concatenate(get(3), get(2)));
  }

  @Override
  public Bytes16 getLow() {
    return Bytes16.wrap(Bytes.concatenate(get(1), get(0)));
  }

  public byte get(final int i, final int j) {
    return get(i).get(j);
  }

  public Bytes getRange(final int i, final int start, final int length) {
    return bytes32.slice(OFFSET * i + start, length);
  }

  public void set(int i, int j, byte b) {
    bytes32.set(OFFSET * i + j, b);
  }

  @Override
  public String toString() {
    final StringBuilder sb = new StringBuilder();
    sb.append("hi=").append(getHigh()).append(", ");
    sb.append("lo=").append(getLow()).append(("\n"));
    sb.append("  0 1 2 3\n  ");
    sb.append(get(0))
        .append(" ")
        .append(get(1))
        .append(" ")
        .append(get(2))
        .append(" ")
        .append(get(3))
        .append(" ");
    return sb.toString();
  }
}
