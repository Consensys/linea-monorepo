package net.consensys.linea.zktracer.module.alu.ext.calculator.mulmod;

import static net.consensys.linea.zktracer.module.Util.isUInt256;
import static net.consensys.linea.zktracer.module.Util.uInt64ToBytes;
import static net.consensys.linea.zktracer.module.alu.ext.calculator.UtilCalculator.calculateProduct;
import static net.consensys.linea.zktracer.module.alu.ext.calculator.UtilCalculator.calculateQuotient;
import static net.consensys.linea.zktracer.module.alu.ext.calculator.UtilCalculator.convertToBaseTheta;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import net.consensys.linea.zktracer.module.alu.ext.BigIntegerConverter;
import org.apache.tuweni.bytes.Bytes32;

public class MulModBytesQCalculator {
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
}
