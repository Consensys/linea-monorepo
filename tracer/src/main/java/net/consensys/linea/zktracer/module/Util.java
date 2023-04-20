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
package net.consensys.linea.zktracer.module;

import static com.google.common.base.Preconditions.checkArgument;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;

public class Util {
  public static Boolean[] byteBits(final UnsignedByte b) {
    final Boolean[] bits = new Boolean[8];
    for (int i = 0; i < 8; i++) {
      bits[7 - i] = b.shiftRight(i).mod(2).toInteger() == 1;
    }
    return bits;
  }

  public static long getOverflow(UInt256 arg, long maxVal, String err) {
    UInt256 shiftedArg = arg.shiftRight(128);

    if (!shiftedArg.fitsLong()) {
      throw new RuntimeException("getOverflow expects a small high part");
    }
    long overflow = shiftedArg.toUnsignedBigInteger().longValue();
    if (overflow > maxVal) {
      throw new RuntimeException(err);
    }
    return overflow;
  }

  // GetBit returns true if the k'th bit of x is 1
  public static boolean getBit(long x, int k) {
    return (x >> k) % 2 == 1;
  }

  public static int boolToInt(boolean bool) {
    return bool ? 1 : 0;
  }

  /**
   * Converts a 64-bit unsigned integer into an 8-byte array
   *
   * @param x The 64-bit unsigned integer to be converted.
   * @return An 8-byte array representing the input value.
   */
  public static byte[] uInt64ToBytes(long x) {
    byte[] xBytes = new byte[8];
    for (int k = 0; k < 8; k++) {
      xBytes[7 - k] = (byte) (x % 256);
      x >>= 8;
    }
    return xBytes;
  }

  public static boolean isUInt256(BigInteger number) {
    return number.bitLength() <= 256;
  }

  public static UInt256 multiplyRange(Bytes[] range1, Bytes[] range2) {
    checkArgument(range1.length == range2.length);
    UInt256 sum = UInt256.ZERO;
    for (int i = 0; i < range1.length; i++) {
      UInt256 prod =
          UInt256.fromBytes(range1[i]).multiply(UInt256.fromBytes(range2[range2.length - i - 1]));
      sum = sum.add(prod);
    }
    return sum;
  }
}
