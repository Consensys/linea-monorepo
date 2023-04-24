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

import net.consensys.linea.zktracer.bytes.UnsignedByte;
import org.apache.tuweni.units.bigints.UInt256;
import org.apache.tuweni.units.bigints.UInt64;

public class Util {

  public static byte boolToByte(boolean b) {
    if (b) {
      return 1;
    }
    return 0;
  }

  public static Boolean[] byteBits(final UnsignedByte b) {
    final Boolean[] bits = new Boolean[8];
    for (int i = 0; i < 8; i++) {
      bits[7 - i] = b.shiftRight(i).mod(2).toInteger() == 1;
    }
    return bits;
  }

  // in Go implementation this method modifies the arg param
  // however (at least in MUL module) the modified value is never used
  // so have not gone to any effort to recreate that behavior in Java implementation
  public static UInt64 getOverflow(final UInt256 arg, final UInt64 maxVal, final String err) {
    UInt256 shifted = arg.shiftRight(128);
    if (shifted.compareTo(UInt64.MAX_VALUE.toBytes()) > 0) {
      // in Go this is panic() but caught by the calling func
      // throw new RuntimeException("getOverflow expects a small high part");
      return UInt64.ZERO;
    }

    UInt64 overflow = UInt64.fromBytes(shifted.trimLeadingZeros());
    if (overflow.compareTo(maxVal) > 0) {
      // in Go this is panic() but caught by the calling func
      // throw new RuntimeException(err + " overflow=" + overflow);
      return UInt64.ZERO;
    }
    return overflow;
  }

  // GetBit returns true iff the k'th bit of x is 1
  public static boolean getBit(UInt64 x, int k) {
    return (x.shiftRight(k)).mod(2).equals(UInt64.ONE);
  }
}
