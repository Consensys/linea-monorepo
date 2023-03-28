package net.consensys.zktracer.module.alu.add;

import net.consensys.zktracer.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

class AdderTest {

    @Test
    void zeroAddZero_isZero() {
        Bytes32 actual = Adder.addSub(OpCode.ADD, Bytes32.ZERO, Bytes32.ZERO);
        assertThat(actual).isEqualTo(Bytes32.ZERO);
    }
    @Test
    void zeroSubZero_isZero() {
        Bytes32 actual = Adder.addSub(OpCode.SUB, Bytes32.ZERO, Bytes32.ZERO);
        assertThat(actual).isEqualTo(Bytes32.ZERO);
    }
    @Test
    void xSubZero_isX() {
        Bytes32 randomBytes = Bytes32.random();
        Bytes32 actual = Adder.addSub(OpCode.SUB, randomBytes, Bytes32.ZERO);
        assertThat(actual).isEqualTo(randomBytes);
    }
    @Test
    void xAddZero_isX() {
        Bytes32 randomBytes = Bytes32.random();
        Bytes32 actual = Adder.addSub(OpCode.ADD, randomBytes, Bytes32.ZERO);
        assertThat(actual).isEqualTo(randomBytes);
    }
}