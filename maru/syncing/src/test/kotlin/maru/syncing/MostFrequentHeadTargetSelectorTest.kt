/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class MostFrequentHeadTargetSelectorTest {
  private val selector = MostFrequentHeadTargetSelector()

  @Test
  fun `should throw exception when peer heads list is empty`() {
    val exception =
      assertThrows<IllegalArgumentException> {
        selector.selectBestSyncTarget(emptyList())
      }
    assertEquals("Peer heads list cannot be empty", exception.message)
  }

  @Test
  fun `should return the only element when list has single item`() {
    val result = selector.selectBestSyncTarget(listOf(100UL))
    assertEquals(100UL, result)
  }

  @Test
  fun `should return most frequent element`() {
    val peerHeads = listOf(100UL, 200UL, 100UL, 300UL, 100UL)
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(100UL, result)
  }

  @Test
  fun `should return highest value when multiple elements have same max frequency`() {
    val peerHeads = listOf(100UL, 200UL, 100UL, 200UL, 300UL)
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(200UL, result) // 100 and 200 both appear twice, return the higher one
  }

  @Test
  fun `should return highest value when all elements appear once`() {
    val peerHeads = listOf(100UL, 200UL, 300UL, 50UL)
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(300UL, result)
  }

  @Test
  fun `should handle large numbers correctly`() {
    val peerHeads =
      listOf(
        ULong.MAX_VALUE,
        ULong.MAX_VALUE - 1UL,
        ULong.MAX_VALUE,
        ULong.MAX_VALUE - 2UL,
      )
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(ULong.MAX_VALUE, result) // MAX_VALUE appears twice
  }

  @Test
  fun `should handle three-way tie by returning highest value`() {
    val peerHeads = listOf(100UL, 200UL, 300UL, 100UL, 200UL, 300UL)
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(300UL, result) // All three values appear twice, return highest
  }

  @Test
  fun `should work with identical elements`() {
    val peerHeads = listOf(500UL, 500UL, 500UL, 500UL)
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(500UL, result)
  }

  @Test
  fun `should handle mixed frequency scenario`() {
    val peerHeads =
      listOf(
        100UL, // appears 1 time
        200UL,
        200UL, // appears 2 times
        300UL,
        300UL,
        300UL, // appears 3 times (most frequent)
        400UL,
        400UL, // appears 2 times
      )
    val result = selector.selectBestSyncTarget(peerHeads)
    assertEquals(300UL, result)
  }
}
