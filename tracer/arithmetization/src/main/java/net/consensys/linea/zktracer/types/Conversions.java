/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.types;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class Conversions {
  public static final Bytes ONE = Bytes.of(1);
  public static final BigInteger UNSIGNED_LONG_MASK =
      BigInteger.ONE.shiftLeft(Long.SIZE).subtract(BigInteger.ONE);

  public static Bytes bigIntegerToBytes(final BigInteger input) {
    Bytes bytes;
    if (input.equals(BigInteger.ZERO)) {
      bytes = Bytes.of(0x00);
    } else {
      byte[] byteArray = input.toByteArray();
      if (byteArray[0] == 0) {
        Bytes tmp = Bytes.wrap(byteArray);
        bytes = Bytes.wrap(tmp.slice(1, tmp.size() - 1));
      } else {
        bytes = Bytes.wrap(byteArray);
      }
    }

    return bytes;
  }

  public static BigInteger unsignedBytesToUnsignedBigInteger(final UnsignedByte[] input) {
    return Bytes.concatenate(Arrays.stream(input).map(i -> Bytes.of(i.toInteger())).toList())
        .toUnsignedBigInteger();
  }

  public static EWord unsignedBytesToEWord(final UnsignedByte[] input) {
    return EWord.of(unsignedBytesToUnsignedBigInteger(input));
  }

  public static UnsignedByte[] bytesToUnsignedBytes(final byte[] bytes) {
    UnsignedByte[] uBytes = new UnsignedByte[bytes.length];
    for (int i = 0; i < bytes.length; i++) {
      uBytes[i] = UnsignedByte.of(bytes[i]);
    }

    return uBytes;
  }

  public static List<UnsignedByte> bytesToUnsignedBytesList(final byte[] bytes) {
    List<UnsignedByte> r = new ArrayList<>(bytes.length);
    for (byte aByte : bytes) {
      r.add(UnsignedByte.of(aByte));
    }

    return r;
  }

  public static BigInteger booleanToBigInteger(final boolean input) {
    return input ? BigInteger.ONE : BigInteger.ZERO;
  }

  public static int booleanToInt(final boolean input) {
    return input ? 1 : 0;
  }
  // Also implemented in oob branch (remove it after merge)
  public static boolean bigIntegerToBoolean(BigInteger n) {
    if (!n.equals(BigInteger.ONE) && !n.equals(BigInteger.ZERO)) {
      throw new IllegalArgumentException(
          "argument should be equal to BigInteger.ONE or BigInteger.ZERO");
    }
    return BigInteger.valueOf(1).equals(n);
  }

  public static BigInteger longToUnsignedBigInteger(final long input) {
    final BigInteger output = BigInteger.valueOf(input).and(UNSIGNED_LONG_MASK);
    if (output.bitLength() > 64) {
      throw new IllegalArgumentException(
          "a long can't be more than 64 bits long, and is" + output.bitLength());
    }
    return output;
  }

  /**
   * Convert the given {@link Bytes} to a signed {@link BigInteger}, even if the Bytes.toBigInteger
   * is overloaded to create unsigned BigInteger (e.g. in {@link
   * org.apache.tuweni.units.bigints.UInt256}.
   *
   * @param a a object implement {@link Bytes}
   * @return the signed BigInteger represented by a's bytes
   */
  public static BigInteger reallyToSignedBigInteger(Bytes a) {
    byte[] bs = a.toArrayUnsafe();
    return new BigInteger(bs, 0, bs.length);
  }

  public static Bytes32 longToBytes32(final long input) {
    return Bytes32.leftPad(Bytes.minimalBytes(input));
  }

  public static Bytes longToBytes(final long input) {
    return input == 0 ? Bytes.of(0) : Bytes.minimalBytes(input);
  }

  public static Bytes booleanToBytes(boolean x) {
    return x ? ONE : Bytes.EMPTY;
  }
}
