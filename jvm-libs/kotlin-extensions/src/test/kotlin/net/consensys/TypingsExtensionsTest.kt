package net.consensys

import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import java.math.BigInteger
import kotlin.random.Random

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
    assertThat((0..0).toIntervalString()).isEqualTo("[0..0]1")
    assertThat((0..1).toIntervalString()).isEqualTo("[0..1]2")
    assertThat((0..10).toIntervalString()).isEqualTo("[0..10]11")
    assertThat((0.0..10.0).toIntervalString()).isEqualTo("[0.0..10.0]11")

    assertThat((0..-1).toIntervalString()).isEqualTo("[0..-1]2")
    assertThat((0..-10).toIntervalString()).isEqualTo("[0..-10]11")
  }

  @Test
  fun `String#decodeHex`() {
    assertThat("0x".decodeHex()).isEmpty()
    assertThat("0x00".decodeHex()).isEqualTo(byteArrayOf(0))
    assertThat("0x01".decodeHex()).isEqualTo(byteArrayOf(1))
    assertThat("0x123456".decodeHex()).isEqualTo(byteArrayOf(0x12, 0x34, 0x56))
  }

  @Test
  fun `ByteArray#encodeHex`() {
    assertThat(byteArrayOf().encodeHex()).isEqualTo("0x")
    assertThat(byteArrayOf(0).encodeHex()).isEqualTo("0x00")
    assertThat(byteArrayOf(1).encodeHex()).isEqualTo("0x01")
    assertThat(byteArrayOf(0x12, 0x34, 0x56).encodeHex()).isEqualTo("0x123456")
  }

  @Test
  fun `ByteArray#assertSize`() {
    assertThatThrownBy {
      byteArrayOf(1, 2, 3).assertSize(2u, "shortNumber")
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("shortNumber expected to have 2 bytes, but got 3")

    assertThat(byteArrayOf(1, 2, 3).assertSize(3u)).isEqualTo(byteArrayOf(1, 2, 3))
  }

  @Test
  fun `ByteArray#assertIs32Bytes`() {
    assertThatThrownBy {
      byteArrayOf(1, 2, 3).assertIs32Bytes("hash")
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("hash expected to have 32 bytes, but got 3")
    ByteArray(32).assertIs32Bytes()
  }

  @Test
  fun `ByteArray#assertIs20Bytes`() {
    assertThatThrownBy {
      byteArrayOf(1, 2, 3).assertIs20Bytes("address")
    }.isInstanceOf(IllegalArgumentException::class.java)
      .hasMessage("address expected to have 20 bytes, but got 3")
    ByteArray(20).assertIs20Bytes()
  }

  @Test
  fun `ByteArray#setFirstByteToZero`() {
    assertThat(Random.Default.nextBytes(32).setFirstByteToZero()[0]).isEqualTo(0)
  }
}
