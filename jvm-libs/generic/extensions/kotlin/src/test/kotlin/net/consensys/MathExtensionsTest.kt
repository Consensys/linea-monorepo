package net.consensys

import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test

class MathExtensionsTest {
  @Test
  fun `UInt addExact()`() {
    assertThat(KMath.addExact(0u, 0u)).isEqualTo(0u)
    assertThat(KMath.addExact(10u, 12u)).isEqualTo(22u)
    assertThat(KMath.addExact(0u, UInt.MAX_VALUE)).isEqualTo(UInt.MAX_VALUE)
    assertThatThrownBy { KMath.addExact(1u, UInt.MAX_VALUE) }
      .isInstanceOf(ArithmeticException::class.java)
      .withFailMessage("UInt overflow")
  }

  @Test
  fun `ULong addExact()`() {
    assertThat(KMath.addExact(0uL, 0uL)).isEqualTo(0uL)
    assertThat(KMath.addExact(10uL, 12uL)).isEqualTo(22uL)
    assertThat(KMath.addExact(0uL, ULong.MAX_VALUE)).isEqualTo(ULong.MAX_VALUE)
    assertThatThrownBy { KMath.addExact(1uL, ULong.MAX_VALUE) }
      .isInstanceOf(ArithmeticException::class.java)
      .withFailMessage("ULong overflow")
  }
}
