package net.consensys.zkevm

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.unsigned.UInt64

class TypingsExtensionsTest {
  @Test
  fun `UInt64#ULong`() {
    assertThat(UInt64.ZERO.toULong()).isEqualTo(0u.toULong())
    assertThat(UInt64.MAX_VALUE.toULong()).isEqualTo(ULong.MAX_VALUE)
    assertThat(UInt64.valueOf(123).toULong()).isEqualTo(123u.toULong())
  }
}
