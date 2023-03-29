package net.consensys.zktracer.module.alu.mul;

import org.junit.jupiter.api.Test;

import java.math.BigInteger;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

public class MulUtilsTest {
    @Test
    public void isTiny() {
        assertThat(MulTracer.isTiny(BigInteger.ZERO)).isTrue();
        assertThat(MulTracer.isTiny(BigInteger.ONE)).isTrue();
        assertThat(MulTracer.isTiny(BigInteger.TWO)).isFalse();
        assertThat(MulTracer.isTiny(BigInteger.TEN)).isFalse();
    }
}
