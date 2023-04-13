package net.consensys.linea.zktracer.module;

import java.math.BigInteger;

import net.consensys.linea.zktracer.bytes.UnsignedByte;
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

  public static int getOverflow(final BigInteger arg, final int maxVal, final String err) {
    BigInteger shiftRight = arg.shiftRight(128);
    if (shiftRight.compareTo(UInt64.MAX_VALUE.toBigInteger()) > 0) {
      throw new RuntimeException("getOverflow expects a small high part");
    }
    int overflow = shiftRight.intValue();
    if (overflow > maxVal) {
      throw new RuntimeException(err);
    }
    return overflow;
  }

  // GetBit returns true iff the k'th bit of x is 1
  public static boolean getBit(int x, int k) {
    return (x >> k) % 2 == 1;
  }
}
