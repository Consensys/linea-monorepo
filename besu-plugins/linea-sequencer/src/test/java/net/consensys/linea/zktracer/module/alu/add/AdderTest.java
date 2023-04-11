package net.consensys.linea.zktracer.module.alu.add;

import static org.assertj.core.api.Assertions.assertThat;

import net.consensys.linea.zktracer.OpCode;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;

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

  @Test
  void maxSubMax_isZero() {
    byte b;
    b = 'f';
    Bytes32 max = Bytes32.repeat(b);
    Bytes32 actual = Adder.addSub(OpCode.SUB, max, max);
    assertThat(actual).isEqualTo(Bytes32.ZERO);
  }

  @Test
  void maxSubZero_isMax() {
    byte b;
    b = 'f';
    Bytes32 max = Bytes32.repeat(b);
    Bytes32 actual = Adder.addSub(OpCode.SUB, max, Bytes32.ZERO);
    assertThat(actual).isEqualTo(max);
  }

  @Test
  void overflowDoesNotError() {
    byte b;
    b = 'f';
    Bytes32 max = Bytes32.repeat(b);
    Adder.addSub(OpCode.ADD, max, max);
  }
}
