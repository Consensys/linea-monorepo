package net.consensys.linea.zktracer.module.alu.ext;

import java.math.BigInteger;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class ExtCalculator {
  private final BigInteger prod;
  long[] prodArray;
  UInt256 arg3;

  public ExtCalculator(Bytes32 arg1, Bytes32 arg2, Bytes32 arg3) {
    BigInteger arg1UInt = arg1.toUnsignedBigInteger();
    BigInteger arg2UInt = arg2.toUnsignedBigInteger();
    prod = arg1UInt.multiply(arg2UInt);
    prodArray = BigIntegerConverter.toLongArray(prod);
    this.arg3 = UInt256.fromBytes(arg3);
  }

  /**
   * Performs extended modular arithmetic operations (MULMOD or ADDMOD) on the given arguments.
   *
   * @param opCode The OpCode representing the modular arithmetic operation (MULMOD or ADDMOD).
   * @param arg1 The first argument for the operation, as a Bytes32 object.
   * @param arg2 The second argument for the operation, as a Bytes32 object.
   * @param arg3 The modulus for the operation, as a Bytes32 object.
   * @return The result of the specified modular arithmetic operation, as a UInt256 object.
   * @throws RuntimeException If an incompatible OpCode is provided for the extended modular
   *     arithmetic module.
   */
  static UInt256 performExtendedModularArithmetic(
      OpCode opCode, Bytes32 arg1, Bytes32 arg2, Bytes32 arg3) {
    // Convert the Bytes32 arguments to UInt256 objects
    UInt256 arg1UInt = UInt256.fromBytes(arg1);
    UInt256 arg2UInt = UInt256.fromBytes(arg2);
    UInt256 arg3UInt = UInt256.fromBytes(arg3);

    // Perform the requested modular arithmetic operation based on the OpCode provided
    switch (opCode) {
        // If the OpCode is MULMOD, perform modular multiplication
      case MULMOD -> {
        return arg1UInt.multiplyMod(arg2UInt, arg3UInt);
      }
        // If the OpCode is ADDMOD, perform modular addition
      case ADDMOD -> {
        return arg1UInt.addMod(arg2UInt, arg3UInt);
      }
        // If an incompatible OpCode is provided, throw a runtime exception
      default -> throw new RuntimeException(
          "Incompatible instruction for extended modular arithmetic module");
    }
  }

  /**
   * Computes and stores delta values and comparison flags for pairs of elements in two arrays. The
   * comparison flag indicates whether the value in cBytes is greater than, less than, or equal to
   * the value in rBytes. The delta value is the absolute difference between the two values, minus
   * one if cBytes is greater than rBytes. The method computes the delta and comparison flag for
   * each pair of elements and stores the delta values as byte arrays in deltaBytes.
   */
  public boolean[] computeComparisonFlags(BaseTheta cBytes, BaseTheta rBytes) {
    boolean[] cmp = new boolean[8];
    for (int i = 0; i < 4; i++) {
      UInt256 c = UInt256.fromBytes(cBytes.get(i));
      UInt256 r = UInt256.fromBytes(rBytes.get(i));
      boolean cGreaterThanR = c.compareTo(r) > 0;
      if (cGreaterThanR) {
        cmp[i] = true;
      } else {
        cmp[4 + i] = c.equals(r);
      }
    }
    return cmp;
  }

  /**
   * Computes and stores delta values and comparison flags for pairs of elements in two arrays. The
   * comparison flag indicates whether the value in cBytes is greater than, less than, or equal to
   * the value in rBytes. The delta value is the absolute difference between the two values, minus
   * one if cBytes is greater than rBytes. The method computes the delta and comparison flag for
   * each pair of elements and stores the delta values as byte arrays in deltaBytes.
   */
  public BaseTheta computeDeltas(BaseTheta cBytes, BaseTheta rBytes) {
    BaseTheta deltaBytes = BaseTheta.fromBytes32(Bytes32.ZERO);
    for (int i = 0; i < 4; i++) {
      UInt256 c = UInt256.fromBytes(cBytes.get(i));
      UInt256 r = UInt256.fromBytes(rBytes.get(i));
      UInt256 delta;

      boolean cGreaterThanR = c.compareTo(r) > 0;
      if (cGreaterThanR) {
        delta = c.subtract(r).subtract(UInt256.ONE);
      } else {
        delta = r.subtract(c);
      }
      // Convert the delta value to a byte array and store it in the ith element of deltaBytes
      BaseTheta truc = (BaseTheta.fromBytes32(delta));
      deltaBytes.set64BitSection(i, truc.get(0));
    }
    return deltaBytes;
  }

  /**
   * Converts the BigInteger array 'prod' into a 8x8 2D byte array.
   *
   * @return A 8x8 2D byte array representing the converted 'prod' BigInteger array.
   */
  public BytesArray computeJs() {
    byte[][] jBytes = new byte[8][8];
    long[] prodArray = BigIntegerConverter.toLongArray(prod);
    for (int k = 0; k < prodArray.length; k++) {
      jBytes[k] = uInt64ToBytes(prodArray[k]);
    }
    return new BytesArray(jBytes);
  }

  public BytesArray computeQs() {
    byte[][] qBytes = new byte[8][8];
    if (isUInt256(prod)) {
      byte[] prodBytes = new byte[32];
      for (int i = 0; i < 4; i++) {
        byte[] bytes = uInt64ToBytes(prodArray[7 - i]);
        System.arraycopy(bytes, 0, prodBytes, i * 8, 8);
      }
      BigInteger prodBigInteger = new BigInteger(1, prodBytes);
      BigInteger quotBigInteger = prodBigInteger.divide(arg3.toUnsignedBigInteger());
      BaseTheta quotBaseTheta = new BaseTheta(UInt256.valueOf(quotBigInteger));
      for (int i = 0; i < 4; i++) {
        qBytes[i] = quotBaseTheta.get(i).toArray();
      }
    } else {
      BigInteger[] divAndRemainder =
          prod.divideAndRemainder(UInt256.fromBytes(arg3).toUnsignedBigInteger());
      long[] quot = BigIntegerConverter.toLongArray(divAndRemainder[0]);
      for (int k = 0; k < quot.length; k++) {
        qBytes[k] = uInt64ToBytes(quot[k]);
      }
    }
    return new BytesArray(qBytes);
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

  private boolean isUInt256(BigInteger number) {
    return number.bitLength() <= 256;
  }
}
