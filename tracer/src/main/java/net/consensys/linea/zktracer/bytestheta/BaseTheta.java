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

package net.consensys.linea.zktracer.bytestheta;

import net.consensys.linea.zktracer.bytes.Bytes16;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

/**
 * The `BaseTheta` class is used to manipulate 256-bit (32-byte) blocks of data, with a focus on
 * dividing the block into four 64-bit (8-byte) sections and performing operations on those
 * sections.
 */
public class BaseTheta extends BaseBytes {

  /**
   * The constructor for the `BaseTheta` class. It takes a parameter of type `Bytes32` called `arg`,
   * which is used to initialize the object.
   *
   * @param arg The `Bytes32` parameter used to initialize the object.
   */
  public BaseTheta(final Bytes32 arg) {
    super(arg);
    bytes32 = arg.mutableCopy();
    for (int k = 0; k < 4; k++) {
      Bytes bytes = arg.slice(OFFSET * k, OFFSET);
      set64BitSection((3 - k), bytes);
    }
  }

  /**
   * This static factory method returns a new instance of the `BaseTheta` class, initialized with
   * the given `arg` parameter.
   *
   * @param arg The `Bytes32` parameter used to initialize the new `BaseTheta` instance.
   * @return A new instance of the `BaseTheta` class, initialized with the given `arg` parameter.
   */
  public static BaseTheta fromBytes32(Bytes32 arg) {
    return new BaseTheta(arg);
  }

  /**
   * This method sets the 64-bit section at the specified `index` in the `bytes32` instance variable
   * to the given `bytes`.
   *
   * @param index The index of the 64-bit section to be set.
   * @param bytes The `Bytes` object that will be used to set the section.
   */
  public void set64BitSection(int index, Bytes bytes) {
    bytes32.set(index * OFFSET, bytes);
  }

  /**
   * This method returns the 64-bit section at the specified `index` in the `bytes32` instance
   * variable.
   *
   * @param index The index of the 64-bit section to be returned.
   * @return The `Bytes` object representing the 64-bit section at the specified index.
   */
  public Bytes get(int index) {
    return bytes32.slice(OFFSET * index, OFFSET);
  }

  /**
   * This method returns a slice of the `bytes32` instance variable, starting at index `i` and
   * extending for `length` bytes.
   *
   * @param i The starting index of the slice.
   * @param length The number of bytes in the slice.
   * @return The `Bytes` object representing the slice of the `bytes32` instance variable.
   */
  public Bytes slice(int i, int length) {
    return bytes32.slice(i, length);
  }

  /**
   * This method returns a new `Bytes16` object that is the concatenation of the third and second
   * 64-bit sections of the `bytes32` instance variable.
   *
   * @return A new `Bytes16` object that is the concatenation of the third and second 64-bit
   *     sections of the `bytes32` instance variable.
   */
  @Override
  public Bytes16 getHigh() {
    return Bytes16.wrap(Bytes.concatenate(get(3), get(2)));
  }

  /**
   * This method returns a new `Bytes16` object that is the concatenation of the first and second
   * 64-bit sections of the `bytes32` instance variable.
   *
   * @return A new `Bytes16` object that is the concatenation of the first and second 64-bit
   *     sections of the `bytes32` instance variable.
   */
  @Override
  public Bytes16 getLow() {
    return Bytes16.wrap(Bytes.concatenate(get(1), get(0)));
  }

  /**
   * This method returns the byte at index `j` in the 64-bit section at index `i` in the `bytes32`
   * instance variable.
   *
   * @param i The index of the 64-bit section to be accessed.
   * @param j The index of the byte within the section to be returned.
   * @return The byte at index `j` in the 64-bit section at index `i` in the `bytes32` instance
   *     variable.
   */
  public byte get(final int i, final int j) {
    return get(i).get(j);
  }

  /**
   * This method returns a slice of the `bytes32` instance variable, starting at index `OFFSET * i +
   * start` and extending for `end` bytes.
   *
   * @param i The index of the 64-bit section to start the slice from.
   * @param start The starting index of the slice within the 64-bit section.
   * @param end The number of bytes in the slice.
   * @return The `Bytes` object representing the slice of the `bytes32` instance variable.
   */
  public Bytes getRange(final int i, final int start, final int end) {
    return bytes32.slice(OFFSET * i + start, end);
  }

  /**
   * This method sets the byte at index `OFFSET * i + j` in the `bytes32` instance variable to the
   * given byte `b`.
   *
   * @param i The index of the 64-bit section to be accessed.
   * @param j The index of the byte within the section to be set.
   * @param b The byte to be set at the specified index.
   */
  public void set(int i, int j, byte b) {
    bytes32.set(OFFSET * i + j, b);
  }

  public Bytes[] getBytesRange(final int start, final int end) {
    int rangeSize = end - start + 1;
    Bytes[] bytes = new Bytes[rangeSize];
    for (int i = 0; i < rangeSize; i++) {
      bytes[i] = get(start + i);
    }
    return bytes;
  }

}
