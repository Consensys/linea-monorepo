package net.consensys.linea.zktracer.module.alu.mul;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import java.math.BigInteger;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.Bytes16;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.module.Util;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.assertj.core.api.Assertions;
import org.junit.jupiter.api.Test;

public class MulUtilsTest {
  @Test
  public void isTiny() {
    // tiny means zero or one
    assertThat(MulData.isTiny(BigInteger.ZERO)).isTrue();
    assertThat(MulData.isTiny(BigInteger.ONE)).isTrue();
    assertThat(MulData.isTiny(BigInteger.TWO)).isFalse();
    assertThat(MulData.isTiny(BigInteger.TEN)).isFalse();
  }

  @Test
  public void twoAdicity() {
    assertThat(MulData.twoAdicity(UInt256.MIN_VALUE)).isEqualTo(256);
    // TODO no idea what these should be
    //    assertThat(MulData.twoAdicity(UInt256.MAX_VALUE)).isEqualTo(0);
    //    assertThat(MulData.twoAdicity(UInt256.valueOf(1))).isEqualTo(0);
  }

  @Test
  public void multiplyByZero() {
    Bytes32 arg1 = Bytes32.random();
    OpCode mul = OpCode.MUL;
    MulData oxo = new MulData(mul, arg1, Bytes32.ZERO);
    Assertions.assertThat(oxo.arg2Hi.isZero()).isTrue();
    Assertions.assertThat(oxo.arg2Lo).isEqualTo(Bytes16.ZERO);
    Assertions.assertThat(oxo.arg2Hi).isEqualTo(Bytes16.ZERO);
    assertThat(oxo.opCode).isEqualTo(mul);
    assertThat(oxo.tinyExponent).isTrue();
    assertThat(oxo.isOneLineInstruction()).isTrue();
    assertThat(oxo.bits[0]).isFalse();
  }

  @Test
  public void zeroExp() {
    Bytes32 arg1 = Bytes32.random();
    OpCode mul = OpCode.EXP;
    MulData oxo = new MulData(mul, arg1, Bytes32.ZERO);
    Assertions.assertThat(oxo.arg2Hi.isZero()).isTrue();
    Assertions.assertThat(oxo.arg2Lo).isEqualTo(Bytes16.ZERO);
    Assertions.assertThat(oxo.arg2Hi).isEqualTo(Bytes16.ZERO);
    assertThat(oxo.opCode).isEqualTo(mul);
    assertThat(oxo.tinyExponent).isTrue();
    assertThat(oxo.isOneLineInstruction()).isTrue();
    assertThat(oxo.bits[0]).isFalse();
  }

  @Test
  public void testByteBits_ofZero() {
    Boolean[] booleans = Util.byteBits(UnsignedByte.of(0));
    assertThat(booleans.length).isEqualTo(8);
    assertThat(booleans[0]).isNotNull();
  }
}
