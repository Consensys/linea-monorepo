/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.kotlin
import maru.extensions.toBytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class ULongExtensionsTest {
  fun String.toByteArray(): ByteArray =
    this
      .removePrefix("0x")
      .chunked(2)
      .map { it.toInt(16).toByte() }
      .toByteArray()

  val uLongParsingTestCases =
    mapOf(
      1UL to "0x0000000000000000000000000000000000000000000000000000000000000001".toByteArray(),
      16UL to "0x0000000000000000000000000000000000000000000000000000000000000010".toByteArray(),
      0xABCDEF1234567890UL to "0x000000000000000000000000000000000000000000000000abcdef1234567890".toByteArray(),
      ULong.MIN_VALUE to "0x0000000000000000000000000000000000000000000000000000000000000000".toByteArray(),
      ULong.MAX_VALUE to "0x000000000000000000000000000000000000000000000000ffffffffffffffff".toByteArray(),
    )

  @Test
  fun toBytes32() {
    uLongParsingTestCases.forEach { (number: ULong, byteArray: ByteArray) ->
      assertThat(number.toBytes32()).isEqualTo(byteArray)
    }
  }
}
