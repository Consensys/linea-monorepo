package net.consensys

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class StringExtensionsTest {
  @Test
  fun `String#decodeHex`() {
    assertThat("0x".decodeHex()).isEmpty()
    assertThat("".decodeHex()).isEmpty()
    assertThat("0x".decodeHex()).isEmpty()
    assertThat("0x00".decodeHex()).isEqualTo(byteArrayOf(0))
    assertThat("0x01".decodeHex()).isEqualTo(byteArrayOf(1))
    assertThat("0x123456".decodeHex()).isEqualTo(byteArrayOf(0x12, 0x34, 0x56))
    assertThat("123456".decodeHex()).isEqualTo(byteArrayOf(0x12, 0x34, 0x56))
  }

  @Test
  fun `String#containsAny`() {
    val stringList = listOf(
      "This is a TEST",
      "lorem ipsum"
    )

    assertThat("this is a test string ignoring cases".containsAny(stringList, ignoreCase = true)).isTrue()
    assertThat("this is a test string without matching cases".containsAny(stringList, ignoreCase = false)).isFalse()
    assertThat("this includes lorem ipsum".containsAny(stringList, ignoreCase = true)).isTrue()
    assertThat("this string won't match".containsAny(stringList, ignoreCase = true)).isFalse()
  }
}
