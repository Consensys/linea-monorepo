package net.consensys.linea.zktracer.module.alu.ext;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.alu.ext.calculator.mulmod.MulModCalculator;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.Test;

public class ExtCalculatorTest {

  @Test
  void performExtendedModularArithmeticMulModTest() {
    Bytes32 arg1 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg2 = fromBigInteger(BigInteger.valueOf(6));
    Bytes32 arg3 = fromBigInteger(BigInteger.valueOf(13));
    UInt256 expected = UInt256.fromBytes(fromBigInteger(BigInteger.valueOf(3)));

    UInt256 result = new MulModCalculator().computeResult(arg1, arg2, arg3);
    assertEquals(expected, result);
  }

  /* @Test
  void performExtendedModularArithmeticAddModTest() {
    Bytes32 arg1 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg2 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg3 = fromBigInteger(BigInteger.valueOf(13));
    UInt256 expected = UInt256.fromBytes(fromBigInteger(BigInteger.valueOf(1)));
    UInt256 result =
        new ExtCalculator(arg1,arg2, arg3).computeResult();
    assertEquals(expected, result);
    assertEquals(expected, result);
  }*/

  private Bytes32 fromBigInteger(BigInteger bigInteger) {
    return Bytes32.leftPad(Bytes.wrap(bigInteger.toByteArray()));
  }
}
