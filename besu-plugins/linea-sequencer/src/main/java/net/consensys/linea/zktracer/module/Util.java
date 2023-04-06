package net.consensys.linea.zktracer.module;

import net.consensys.linea.zktracer.bytes.UnsignedByte;

public class Util {
  public static Boolean[]  byteBits(final UnsignedByte b) {
    final Boolean[] bits = new Boolean[8];
    for (int i = 0; i < 8; i++) {
      bits[7 - i] = b.shiftRight(i).mod(2).toInteger() == 1;
    }
    return bits;
  }
}
