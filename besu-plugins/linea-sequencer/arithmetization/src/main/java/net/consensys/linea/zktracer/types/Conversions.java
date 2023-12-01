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
import java.util.Arrays;
import java.util.stream.IntStream;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class Conversions {
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
    return (UnsignedByte[])
        IntStream.range(0, bytes.length).mapToObj(i -> UnsignedByte.of(bytes[i])).toArray();
  }

  public static BigInteger booleanToBigInteger(final boolean input) {
    return input ? BigInteger.ONE : BigInteger.ZERO;
  }

  public static BigInteger longToUnsignedBigInteger(final long input) {
    final BigInteger output = BigInteger.valueOf(input).and(UNSIGNED_LONG_MASK);
    if (output.bitLength() > 64) {
      throw new IllegalArgumentException(
          "a long can't be more than 64 bits long, and is" + output.bitLength());
    }
    return output;
  }

  public static Bytes32 longToBytes32(final long input) {
    return Bytes32.leftPad(Bytes.minimalBytes(input));
  }

  public static Bytes longToBytes(final long input) {
    return input == 0 ? Bytes.of(0) : Bytes.minimalBytes(input);
  }
}
