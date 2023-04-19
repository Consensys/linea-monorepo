package net.consensys.linea.zktracer.module.alu.ext;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.math.BigInteger;

import net.consensys.linea.zktracer.OpCode;
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

    UInt256 result =
        ExtCalculator.performExtendedModularArithmetic(OpCode.MULMOD, arg1, arg2, arg3);
    assertEquals(expected, result);
  }

  @Test
  void performExtendedModularArithmeticAddModTest() {
    Bytes32 arg1 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg2 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg3 = fromBigInteger(BigInteger.valueOf(13));
    UInt256 expected = UInt256.fromBytes(fromBigInteger(BigInteger.valueOf(1)));
    UInt256 result =
        ExtCalculator.performExtendedModularArithmetic(OpCode.ADDMOD, arg1, arg2, arg3);
    assertEquals(expected, result);
    assertEquals(expected, result);
  }

  @Test
  void performExtendedModularArithmeticInvalidOpCodeTest() {
    // Test case: invalid OpCode
    Bytes32 arg1 = fromBigInteger(BigInteger.valueOf(7));
    Bytes32 arg2 = fromBigInteger(BigInteger.valueOf(6));
    Bytes32 arg3 = fromBigInteger(BigInteger.valueOf(13));

    assertThrows(
        RuntimeException.class,
        () -> {
          ExtCalculator.performExtendedModularArithmetic(OpCode.SLT, arg1, arg2, arg3);
        });
  }

  private Bytes32 fromBigInteger(BigInteger bigInteger) {
    return Bytes32.leftPad(Bytes.wrap(bigInteger.toByteArray()));
  }
}
