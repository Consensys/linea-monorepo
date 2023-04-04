package net.consensys.zktracer.module.alu.mod;

import java.math.BigInteger;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class BaseTheta {
  final private Bytes[] bytes  = new Bytes[4];
  BaseTheta(final BigInteger arg) {
    Bytes32 t = Bytes32.leftPad(Bytes.wrap(arg .toByteArray()));
    for (int k = 0; k < 4; k++) {
      bytes[3-k] = t.slice(8*k, 8);
    }
  }

  public Bytes get(int index) {
    return bytes[index];
  }
}
