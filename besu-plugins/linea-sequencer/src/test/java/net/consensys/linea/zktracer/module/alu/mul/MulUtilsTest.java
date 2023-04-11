package net.consensys.linea.zktracer.module.alu.mul;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import java.math.BigInteger;

import org.junit.jupiter.api.Test;

public class MulUtilsTest {
  @Test
  public void isTiny() {
    assertThat(MulTracer.isTiny(BigInteger.ZERO)).isTrue();
    assertThat(MulTracer.isTiny(BigInteger.ONE)).isTrue();
    assertThat(MulTracer.isTiny(BigInteger.TWO)).isFalse();
    assertThat(MulTracer.isTiny(BigInteger.TEN)).isFalse();
  }
}
