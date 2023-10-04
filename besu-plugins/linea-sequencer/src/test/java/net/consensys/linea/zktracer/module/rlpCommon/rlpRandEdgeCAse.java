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

package net.consensys.linea.zktracer.module.rlpCommon;

import java.math.BigInteger;
import java.util.Random;

import org.apache.tuweni.bytes.Bytes;

public class rlpRandEdgeCAse {
  private static Random rnd = new Random(666);

  public static final BigInteger randBigInt(boolean onlySixteenByte) {
    int selectorBound = 4;
    if (!onlySixteenByte) {
      selectorBound += 1;
    }
    int selector = rnd.nextInt(0, selectorBound);

    return switch (selector) {
      case 0 -> BigInteger.ZERO;
      case 1 -> BigInteger.valueOf(rnd.nextInt(1, 128));
      case 2 -> BigInteger.valueOf(rnd.nextInt(128, 256));
      case 3 -> new BigInteger(16 * 8, rnd);
      case 4 -> new BigInteger(32 * 8, rnd);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }

  public static Bytes randData(boolean nonEmpty) {
    int selectorOrigin = 0;
    if (nonEmpty) {
      selectorOrigin += 1;
    }
    int selector = rnd.nextInt(selectorOrigin, 6);
    return switch (selector) {
      case 0 -> Bytes.EMPTY;
      case 1 -> Bytes.of(0x0);
      case 2 -> Bytes.minimalBytes(rnd.nextLong(1, 128));
      case 3 -> Bytes.minimalBytes(rnd.nextLong(128, 256));
      case 4 -> Bytes.random(rnd.nextInt(1, 56), rnd);
      case 5 -> Bytes.random(rnd.nextInt(56, 666), rnd);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }

  public static Long randLong() {
    int selector = rnd.nextInt(0, 4);
    return switch (selector) {
      case 0 -> 0L;
      case 1 -> rnd.nextLong(1, 128);
      case 2 -> rnd.nextLong(128, 256);
      case 3 -> rnd.nextLong(256, 0xfffffffffffffffL);
      default -> throw new IllegalStateException("Unexpected value: " + selector);
    };
  }
}
