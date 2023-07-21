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

package net.consensys.linea.zktracer.bytes;

import java.util.Arrays;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.MutableBytes;

class MutableArrayWrappingBytes extends ArrayWrappingBytes implements MutableBytes {

  MutableArrayWrappingBytes(byte[] bytes) {
    super(bytes);
  }

  MutableArrayWrappingBytes(byte[] bytes, int offset, int length) {
    super(bytes, offset, length);
  }

  @Override
  public void set(int i, byte b) {
    // Check bounds because while the array access would throw, the error message
    // would be confusing
    // for the caller.
    Checks.checkElementIndex(i, length);
    bytes[offset + i] = b;
  }

  @Override
  public void set(int i, Bytes b) {
    byte[] bytesArray = b.toArrayUnsafe();
    System.arraycopy(bytesArray, 0, bytes, offset + i, bytesArray.length);
  }

  @Override
  public MutableBytes increment() {
    for (int i = length - 1; i >= offset; --i) {
      if (bytes[i] == (byte) 0xFF) {
        bytes[i] = (byte) 0x00;
      } else {
        ++bytes[i];
        break;
      }
    }

    return this;
  }

  @Override
  public MutableBytes decrement() {
    for (int i = length - 1; i >= offset; --i) {
      if (bytes[i] == (byte) 0x00) {
        bytes[i] = (byte) 0xFF;
      } else {
        --bytes[i];
        break;
      }
    }

    return this;
  }

  @Override
  public MutableBytes mutableSlice(int i, int length) {
    if (i == 0 && length == this.length) {
      return this;
    }
    if (length == 0) {
      return MutableBytes.EMPTY;
    }

    Checks.checkElementIndex(i, this.length);
    Checks.checkArgument(
        i + length <= this.length,
        "Specified length %s is too large: the value has size %s and has only %s bytes from %s",
        length,
        this.length,
        this.length - i,
        i);

    return length == Bytes16.SIZE
        ? new MutableArrayWrappingBytes16(bytes, offset + i)
        : new MutableArrayWrappingBytes(bytes, offset + i, length);
  }

  @Override
  public void fill(byte b) {
    Arrays.fill(bytes, offset, offset + length, b);
  }

  @Override
  public Bytes copy() {
    return new ArrayWrappingBytes(toArray());
  }

  @Override
  public int hashCode() {
    return computeHashcode();
  }
}
