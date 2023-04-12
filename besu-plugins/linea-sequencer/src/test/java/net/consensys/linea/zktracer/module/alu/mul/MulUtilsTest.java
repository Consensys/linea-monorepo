package net.consensys.linea.zktracer.module.alu.mul;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import java.math.BigInteger;

import org.apache.tuweni.units.bigints.UInt256;
import org.junit.jupiter.api.Test;

public class MulUtilsTest {
  @Test
  public void isTiny() {
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
}
