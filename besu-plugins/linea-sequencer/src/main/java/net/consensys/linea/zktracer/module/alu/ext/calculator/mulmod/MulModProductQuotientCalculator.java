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
package net.consensys.linea.zktracer.module.alu.ext.calculator.mulmod;

import static net.consensys.linea.zktracer.module.Util.isUInt256;
import static net.consensys.linea.zktracer.module.Util.uInt64ToBytes;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import net.consensys.linea.zktracer.module.alu.ext.BigIntegerConverter;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

/*
  This class provides methods for computing the product and quotient of three 32-byte arguments,
  and returning the result as a BytesArray. (MulMod)
*/
public class MulModProductQuotientCalculator {
  /**
   * Computes the product of two Bytes32 arguments and returns the result as a BytesArray.
   *
   * @param arg1 The first Bytes32 argument.
   * @param arg2 The second Bytes32 argument.
   * @return The product of the two arguments as a BytesArray.
   */
  public static BytesArray computeJs(Bytes32 arg1, Bytes32 arg2) {
    byte[][] jBytes = new byte[8][8];
    BigInteger arg1UInt = arg1.toUnsignedBigInteger();
    BigInteger arg2UInt = arg2.toUnsignedBigInteger();
    // Compute the product of the two unsigned BigIntegers.
    BigInteger prod = arg1UInt.multiply(arg2UInt);

    long[] prodArray = BigIntegerConverter.toLongArray(prod);
    // Iterate through the long array and convert each element to a byte array.
    for (int k = 0; k < prodArray.length; k++) {
      jBytes[k] = uInt64ToBytes(prodArray[k]);
    }
    return new BytesArray(jBytes);
  }

  /**
   * Computes the quotient of the product of arg1 and arg2 divided by arg3, all of Bytes32 type, and
   * returns the result as a BytesArray.
   *
   * @param arg1 The first Bytes32 argument.
   * @param arg2 The second Bytes32 argument.
   * @param arg3 The third Bytes32 argument.
   * @return The quotient of the product of arg1 and arg2 divided by arg3 as a BytesArray.
   */
  public static BytesArray computeQs(Bytes32 arg1, Bytes32 arg2, Bytes32 arg3) {
    byte[][] qBytes = new byte[8][8];

    BigInteger prod = calculateProduct(arg1, arg2);

    if (isUInt256(prod)) {
      BigInteger quotBigInteger = calculateQuotient(prod, arg3);
      BaseTheta quotBaseTheta = convertToBaseTheta(quotBigInteger);

      for (int i = 0; i < 4; i++) {
        // Copy the BaseTheta byte arrays into the result byte array.
        qBytes[i] = quotBaseTheta.get(i).toArray();
      }
    } else {
      // Divide the product by arg3 and convert the quotient to a byte array.
      BigInteger[] divAndRemainder = prod.divideAndRemainder(arg3.toUnsignedBigInteger());
      long[] quot = BigIntegerConverter.toLongArray(divAndRemainder[0]);
      for (int k = 0; k < quot.length; k++) {
        qBytes[k] = uInt64ToBytes(quot[k]);
      }
    }

    // Return the result as a BytesArray.
    return new BytesArray(qBytes);
  }

  /**
   * Calculates the product of arg1 and arg2, both of type Bytes32, and returns the result as a
   * BigInteger.
   *
   * @param arg1 The first Bytes32 argument.
   * @param arg2 The second Bytes32 argument.
   * @return The product of arg1 and arg2 as a BigInteger.
   */
  private static BigInteger calculateProduct(Bytes32 arg1, Bytes32 arg2) {
    BigInteger arg1UInt = arg1.toUnsignedBigInteger();
    BigInteger arg2UInt = arg2.toUnsignedBigInteger();

    // Compute the product of the two unsigned BigIntegers.
    return arg1UInt.multiply(arg2UInt);
  }

  /**
   * Calculates the quotient of the given BigInteger divided by arg3, a Bytes32 argument. Returns
   * the result as a BigInteger.
   *
   * @param prod The BigInteger to divide.
   * @param arg3 The third Bytes32 argument.
   * @return The quotient of the given BigInteger and arg3 as a BigInteger.
   */
  private static BigInteger calculateQuotient(BigInteger prod, Bytes32 arg3) {
    byte[] prodBytes = prod.toByteArray(); // Convert the product to a byte array.
    return new BigInteger(1, prodBytes).divide(arg3.toUnsignedBigInteger());
  }

  /**
   * Converts a given BigInteger to a BaseTheta representation.
   *
   * @param quotBigInteger The BigInteger to convert.
   * @return The given BigInteger as a BaseTheta representation.
   */
  private static BaseTheta convertToBaseTheta(BigInteger quotBigInteger) {
    return new BaseTheta(UInt256.valueOf(quotBigInteger));
  }
}
