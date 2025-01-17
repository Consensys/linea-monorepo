package net.consensys

import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
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
  fun `toHexStringPaddedToBitSize_shouldPadToDesireNumberOfBitsWhenItFits`() {
    assertThat(0.toULong().toHexStringPaddedToBitSize(4)).isEqualTo("0x0")
    assertThat(0.toULong().toHexStringPaddedToBitSize(8)).isEqualTo("0x00")
    assertThat(1.toULong().toHexStringPaddedToBitSize(4)).isEqualTo("0x1")
    assertThat(1.toULong().toHexStringPaddedToBitSize(12)).isEqualTo("0x001")
    assertThat(255.toULong().toHexStringPaddedToBitSize(8)).isEqualTo("0xff")

    assertThat(ULong.MAX_VALUE.toHexStringPaddedToBitSize(64))
      .isEqualTo("0xffffffffffffffff")
    assertThat(ULong.MAX_VALUE.toHexStringPaddedToBitSize(80))
      .isEqualTo("0x0000ffffffffffffffff")
  }

  @Test
  fun `toHexStringPaddedToBitSize_shouldThrowErrorWhenBitSizeIsNotMultipleOf4`() {
    assertThatThrownBy { 2.toULong().toHexStringPaddedToBitSize(9) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("targetBitSize=9 should be a multiple of 4")
  }

  @Test
  fun `toHexStringPaddedToBitSize_shouldThrowErrorWhenNumberDoesNotFitWithTargetNumberOfBits`() {
    assertThatThrownBy { 256.toULong().toHexStringPaddedToBitSize(8) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("Number 256 needs 3 hex digits (12 bits), targetBitSize=8")
  }

  @Test
  fun `toHexStringPaddedToByteSize_shouldPadToDesireNumberOfBitsWhenItFits`() {
    assertThat(0.toULong().toHexStringPaddedToByteSize(1)).isEqualTo("0x00")
    assertThat(1.toULong().toHexStringPaddedToByteSize(1)).isEqualTo("0x01")
    assertThat(2.toULong().toHexStringPaddedToByteSize(3)).isEqualTo("0x000002")
    assertThat(255.toULong().toHexStringPaddedToByteSize(1)).isEqualTo("0xff")
    assertThat(ULong.MAX_VALUE.toHexStringPaddedToByteSize(8))
      .isEqualTo("0xffffffffffffffff")
    assertThat(ULong.MAX_VALUE.toHexStringPaddedToByteSize(10))
      .isEqualTo("0x0000ffffffffffffffff")
  }

  @Test
  fun `toHexStringPaddedToByteSize_shouldThrowErrorWhenNumberDoesNotFitWithTargetNumber`() {
    assertThatThrownBy { 256.toULong().toHexStringPaddedToByteSize(1) }
      .isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("Number 256 needs 3 hex digits (12 bits), targetBitSize=8")
  }

  @Test
  fun `toHexStringUInt256_shouldPadTo256Bit`() {
    assertThat(0.toULong().toHexStringUInt256()).isEqualTo("0x" + "0".repeat(64))
    assertThat(15.toULong().toHexStringUInt256()).isEqualTo("0x" + "0".repeat(63) + "f")
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
    assertThat((0..0).toIntervalString()).isEqualTo("[0..0]1")
    assertThat((0..1).toIntervalString()).isEqualTo("[0..1]2")
    assertThat((0..10).toIntervalString()).isEqualTo("[0..10]11")
    assertThat((0.0..10.0).toIntervalString()).isEqualTo("[0.0..10.0]11")

    assertThat((0..-1).toIntervalString()).isEqualTo("[0..-1]2")
    assertThat((0..-10).toIntervalString()).isEqualTo("[0..-10]11")
  }
}
