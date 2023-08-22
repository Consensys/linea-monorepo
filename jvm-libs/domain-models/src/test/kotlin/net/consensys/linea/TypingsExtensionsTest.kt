package net.consensys.linea

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import java.math.BigInteger

class TypingsExtensionsTest {

  private val uLongParsingTestCases = mapOf(
    0.toULong() to "0x0",
    1.toULong() to "0x1",
    0xABC_DEF_123_456u to "0xabcdef123456",
    ULong.MAX_VALUE to "0xffffffffffffffff"
  )

  @Test
  fun `ULong toHexString`() {
    uLongParsingTestCases.forEach { (number: ULong, hexString: String) ->
      assertThat(number.toHexString()).isEqualTo(hexString)
    }
  }

  @Test
  fun `ULong fromHexString`() {
    uLongParsingTestCases.forEach { (number: ULong, hexString: String) ->
      assertThat(ULong.fromHexString(hexString)).isEqualTo(number)
    }
  }

  @Test
  fun `ULong fromHexString invalid format`() {
    assertThrows<NumberFormatException> { ULong.fromHexString("0x23J") }
  }

  @Test
  fun `BigInteger#ULong`() {
    assertThrows<NumberFormatException> { BigInteger.valueOf(-1).toULong() }
    assertThat(BigInteger.valueOf(0).toULong()).isEqualTo(0u.toULong())
    assertThat(BigInteger.valueOf(123).toULong()).isEqualTo(123u.toULong())
    assertThat(BigInteger(ULong.MAX_VALUE.toString(), 10).toULong()).isEqualTo(ULong.MAX_VALUE)
  }

  @Test
  fun `ULong#BigInteger`() {
    assertThat(0UL.toBigInteger()).isEqualTo(BigInteger.valueOf(0))
    assertThat(123UL.toBigInteger()).isEqualTo(BigInteger.valueOf(123))
    assertThat(ULong.MAX_VALUE.toBigInteger()).isEqualTo(BigInteger(ULong.MAX_VALUE.toString(), 10))
  }

  @Test
  fun `toIntervalString`() {
    assertThat((0..0).toIntervalString()).isEqualTo("[0..0](1)")
    assertThat((0..1).toIntervalString()).isEqualTo("[0..1](2)")
    assertThat((0..10).toIntervalString()).isEqualTo("[0..10](11)")
    assertThat((0.0..10.0).toIntervalString()).isEqualTo("[0.0..10.0](11.0)")

    assertThat((0..-1).toIntervalString()).isEqualTo("[0..-1](2)")
    assertThat((0..-10).toIntervalString()).isEqualTo("[0..-10](11)")
  }
}
