package net.consensys.tuweni.bytes

import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import kotlin.random.Random

class Bytes32Test {
  @Test
  fun sliceAsBytes32() {
    val bytes = Random.Default.nextBytes(3 * 32 - 1)
    assertThat(bytes.sliceAsBytes32(0)).isEqualTo(Bytes32.wrap(bytes, 0))
    assertThat(bytes.sliceAsBytes32(1)).isEqualTo(Bytes32.wrap(bytes, 32))
    assertThatThrownBy { bytes.sliceAsBytes32(2) }
      .isInstanceOf(IllegalArgumentException::class.java)
  }
}
