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
package net.consensys.linea.zktracer.module.alu.ext.calculator.addmod;

import static net.consensys.linea.zktracer.module.Util.uInt64ToBytes;
import static net.consensys.linea.zktracer.module.alu.ext.BigIntegerConverter.fromLongArray;
import static net.consensys.linea.zktracer.module.alu.ext.calculator.UtilCalculator.calculateQuotient;
import static net.consensys.linea.zktracer.module.alu.ext.calculator.UtilCalculator.calculateSum;
import static net.consensys.linea.zktracer.module.alu.ext.calculator.UtilCalculator.convertToBaseTheta;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import net.consensys.linea.zktracer.module.alu.ext.BigIntegerConverter;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class AddModBytesQCalculator {
  public static BytesArray computeQs(Bytes32 arg1, Bytes32 arg2, Bytes32 arg3) {
    byte[][] qBytes = new byte[8][8];
    UInt256 sum = calculateSum(arg1, arg2);

    // Addition does not overflow
    if (UInt256.fromBytes(arg1).compareTo(sum) <= 0) {
      BigInteger quotBigInteger = calculateQuotient(sum.toUnsignedBigInteger(), arg3);
      BaseTheta quotBaseTheta = convertToBaseTheta(quotBigInteger);

      for (int k = 0; k < 4; k++) {
        qBytes[k] = quotBaseTheta.get(k).toArray();
      }
    } else {
      BigInteger bigIntegerOverflow = convertToBigIntegerWithOverflow(sum);
      BigInteger[] divAndRemainder =
          bigIntegerOverflow.divideAndRemainder(arg3.toUnsignedBigInteger());
      long[] quot = BigIntegerConverter.toLongArray(divAndRemainder[0]);
      for (int k = 0; k < quot.length; k++) {
        qBytes[k] = uInt64ToBytes(quot[k]);
      }
    }
    return new BytesArray(qBytes);
  }

  private static BigInteger convertToBigIntegerWithOverflow(UInt256 sum) {

    long[] sumUInt64 = new long[5];
    for (int k = 0; k < 4; k++) {
      sumUInt64[k] = sum.toUnsignedBigInteger().shiftRight(k * 64).longValue();
    }
    sumUInt64[4] = 1L;
    return fromLongArray(sumUInt64);
  }
}
