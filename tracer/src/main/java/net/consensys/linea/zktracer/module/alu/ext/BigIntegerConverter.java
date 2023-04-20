package net.consensys.linea.zktracer.module.alu.ext;

import java.math.BigInteger;
import java.nio.ByteBuffer;

public class BigIntegerConverter {

  public static long[] toLongArray(BigInteger bigInteger) {
    // Ensure the input BigInteger is within 512 bits
    if (bigInteger.bitLength() > 512) {
      throw new IllegalArgumentException("BigInteger cannot be larger than 512 bits.");
    }

    // Convert the BigInteger to a byte array and pad it to 64 bytes (512 bits)
    byte[] byteArray = new byte[64];
    byte[] inputBytes = bigInteger.toByteArray();
    int start = byteArray.length - inputBytes.length;
    System.arraycopy(inputBytes, 0, byteArray, start, inputBytes.length);

    // Convert the byte array to an array of 8 longs
    long[] longArray = new long[8];
    ByteBuffer buffer = ByteBuffer.wrap(byteArray);
    for (int i = 0; i < 8; i++) {
      longArray[7 - i] = buffer.getLong(i * 8);
    }
    return longArray;
  }
}
