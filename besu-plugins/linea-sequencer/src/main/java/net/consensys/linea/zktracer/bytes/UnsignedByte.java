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

import com.fasterxml.jackson.annotation.JsonValue;

public class UnsignedByte {
  private final short unsignedByte;

  private UnsignedByte(final short unsignedByte) {
    this.unsignedByte = unsignedByte;
  }

  public static UnsignedByte of(final byte b) {
    final short unsignedB = (short) (b & 0xff);

    checkLength(unsignedB);

    return new UnsignedByte(unsignedB);
  }

  public static UnsignedByte of(final int b) {
    checkLength(b);

    return new UnsignedByte((short) b);
  }

  public static UnsignedByte of(final long b) {
    checkLength(b);

    return new UnsignedByte((short) b);
  }

  public UnsignedByte shiftLeft(final UnsignedByte shiftAmount) {
    return shiftLeft(shiftAmount.toInteger());
  }

  public UnsignedByte shiftLeft(final int shiftAmount) {
    return new UnsignedByte((short) ((unsignedByte << shiftAmount) & 0xff));
  }

  public UnsignedByte shiftRight(final UnsignedByte shiftAmount) {
    return shiftRight(shiftAmount.toInteger());
  }

  public UnsignedByte shiftRight(final int shiftAmount) {
    return new UnsignedByte((short) ((unsignedByte >> shiftAmount) & 0xff));
  }

  public UnsignedByte mod(final int m) {
    return new UnsignedByte((short) (unsignedByte % m));
  }

  @JsonValue
  public int toInteger() {
    return unsignedByte;
  }

  private static void checkLength(final long b) {
    if (b < 0 || b > 255) {
      throw new IllegalArgumentException("Unsigned byte value must be between 0 - 255. Is " + b);
    }
  }
}
