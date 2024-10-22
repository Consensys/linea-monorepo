package build.linea.tuweni

import net.consensys.toBigInteger
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import kotlin.random.Random

class Bytes32Test {
  @BeforeEach
  fun setUp() {
    // workaround: need this to load the functions otherwise JUNit gets stuck ¯\_(ツ)_/¯
    Random.Default.nextBytes(3 * 32 - 1).sliceAsBytes32(0)
    UInt256.ZERO.toBytes().toULong()
  }

  @Test
  fun testSliceAsBytes32() {
    val bytes = Random.Default.nextBytes(3 * 32 - 1)
    assertThat(bytes.sliceAsBytes32(0)).isEqualTo(Bytes32.wrap(bytes, 0))
    assertThat(bytes.sliceAsBytes32(1)).isEqualTo(Bytes32.wrap(bytes, 32))
    assertThatThrownBy { bytes.sliceAsBytes32(2) }
      .isInstanceOf(IllegalArgumentException::class.java)
  }

  @Test
  fun testToULong() {
    assertThat(UInt256.ZERO.toBytes().toULong()).isEqualTo(0uL)
    assertThat(UInt256.valueOf(Long.MAX_VALUE).toULong()).isEqualTo(Long.MAX_VALUE.toULong())
    assertThat(UInt256.valueOf(Long.MAX_VALUE).add(UInt256.ONE).toULong()).isEqualTo(Long.MAX_VALUE.toULong() + 1UL)
    assertThat(UInt256.valueOf(ULong.MAX_VALUE.toBigInteger()).toULong()).isEqualTo(ULong.MAX_VALUE)
  }
}
