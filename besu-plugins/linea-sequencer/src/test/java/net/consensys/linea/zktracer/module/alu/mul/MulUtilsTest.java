package net.consensys.linea.zktracer.module.alu.mul;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import java.math.BigInteger;

import org.junit.jupiter.api.Test;

public class MulUtilsTest {
  @Test
  public void isTiny() {
    assertThat(MulData.isTiny(BigInteger.ZERO)).isTrue();
    assertThat(MulData.isTiny(BigInteger.ONE)).isTrue();
    assertThat(MulData.isTiny(BigInteger.TWO)).isFalse();
    assertThat(MulData.isTiny(BigInteger.TEN)).isFalse();
  }
}
