/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class ForksScheduleTest {
  private val consensusConfig = object : ConsensusConfig {}
  private val qbftConsensusConfig = object : ConsensusConfig {}
  private val otherConsensusConfig = object : ConsensusConfig {}
  private val expectedChainId = 1337u

  @Test
  fun `throws exception on duplicate timestamps`() {
    val fork1 = ForkSpec(timestampSeconds = 1L, blockTimeSeconds = 1, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 1L, blockTimeSeconds = 2, configuration = consensusConfig)
    val fork3 = ForkSpec(timestampSeconds = 3L, blockTimeSeconds = 3, configuration = consensusConfig)
    val forks = listOf(fork1, fork2, fork3)

    val exception = assertThrows<IllegalArgumentException> { ForksSchedule(expectedChainId, forks) }
    assertThat(exception).hasMessageContaining("Fork timestamps must be unique")
  }

  @Test
  fun `test getForkByTimestamp returns correct fork`() {
    val fork1 = ForkSpec(timestampSeconds = 1L, blockTimeSeconds = 1, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2L, blockTimeSeconds = 2, configuration = consensusConfig)
    val fork3 = ForkSpec(timestampSeconds = 3L, blockTimeSeconds = 3, configuration = consensusConfig)
    val forks = listOf(fork1, fork2, fork3)

    val schedule = ForksSchedule(expectedChainId, forks)

    assertThat(schedule.getForkByTimestamp(1L)).isEqualTo(fork1)
    assertThat(schedule.getForkByTimestamp(2L)).isEqualTo(fork2)
    assertThat(schedule.getForkByTimestamp(3L)).isEqualTo(fork3)
  }

  @Test
  fun `getForkByTimestamp throws if timestamp is before all forks`() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = consensusConfig)
    val forks = listOf(fork1, fork2)

    val schedule = ForksSchedule(expectedChainId, forks)

    val exception =
      assertThrows<IllegalArgumentException> {
        schedule.getForkByTimestamp(500L)
      }
    assertThat(exception).hasMessageContaining("No fork found")
  }

  @Test
  fun `ForkSpec initialization with invalid blockTimeSeconds`() {
    val exception =
      assertThrows<IllegalArgumentException> {
        ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 0, configuration = consensusConfig)
      }
    assertThat(exception).hasMessage("blockTimeSeconds must be greater or equal to 1 second")
  }

  @Test
  fun equality() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = consensusConfig)
    val forks1 = listOf(fork1, fork2)
    val forks2 = listOf(fork1, fork2)

    val schedule1 = ForksSchedule(expectedChainId, forks1)
    val schedule2 = ForksSchedule(expectedChainId, forks2)

    assertThat(schedule1).isEqualTo(schedule2)
    assertThat(schedule1.hashCode()).isEqualTo(schedule2.hashCode())
  }

  @Test
  fun inequality() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = consensusConfig)
    val fork3 = ForkSpec(timestampSeconds = 3000L, blockTimeSeconds = 30, configuration = consensusConfig)
    val forks1 = listOf(fork1, fork2)
    val forks2 = listOf(fork1, fork3)

    val schedule1 = ForksSchedule(expectedChainId, forks1)
    val schedule2 = ForksSchedule(expectedChainId, forks2)

    assertThat(schedule1).isNotEqualTo(schedule2)
    assertThat(schedule1.hashCode()).isNotEqualTo(schedule2.hashCode())
  }

  @Test
  fun `getForkByConfigType throws exception when config class not found`() {
    val qbftFork = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = qbftConsensusConfig)
    val forks = listOf(qbftFork)

    val schedule = ForksSchedule(expectedChainId, forks)

    val exception =
      assertThrows<IllegalArgumentException> {
        schedule.getForkByConfigType(otherConsensusConfig::class)
      }
    assertThat(exception).hasMessageContaining("No fork found for config type")
  }

  @Test
  fun `getForkByConfigType returns first matching fork`() {
    val otherFork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val qbftFork1 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = qbftConsensusConfig)
    val otherFork2 = ForkSpec(timestampSeconds = 3000L, blockTimeSeconds = 30, configuration = consensusConfig)
    val qbftFork2 = ForkSpec(timestampSeconds = 4000L, blockTimeSeconds = 40, configuration = qbftConsensusConfig)
    val qbftFork3 = ForkSpec(timestampSeconds = 5000L, blockTimeSeconds = 50, configuration = qbftConsensusConfig)
    val forks = listOf(otherFork1, qbftFork1, otherFork2, qbftFork2, qbftFork3)

    val schedule = ForksSchedule(expectedChainId, forks)

    // Should return the first one found (note: forks are sorted by timestamp in reverse order internally)
    val result = schedule.getForkByConfigType(qbftConsensusConfig::class)
    assertThat(result).isEqualTo(qbftFork1)
  }

  @Test
  fun `getNextForkByTimestamp returns next fork when one exists`() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = qbftConsensusConfig)
    val fork3 = ForkSpec(timestampSeconds = 3000L, blockTimeSeconds = 30, configuration = otherConsensusConfig)
    val forks = listOf(fork1, fork2, fork3)

    val schedule = ForksSchedule(expectedChainId, forks)

    // Test getting next fork from before first fork
    assertThat(schedule.getNextForkByTimestamp(500L)).isEqualTo(fork1)

    // Test getting next fork from first fork timestamp
    assertThat(schedule.getNextForkByTimestamp(1000L)).isEqualTo(fork2)

    // Test getting next fork from between first and second fork
    assertThat(schedule.getNextForkByTimestamp(1500L)).isEqualTo(fork2)

    // Test getting next fork from second fork timestamp
    assertThat(schedule.getNextForkByTimestamp(2000L)).isEqualTo(fork3)

    // Test getting next fork from between second and third fork
    assertThat(schedule.getNextForkByTimestamp(2500L)).isEqualTo(fork3)
  }

  @Test
  fun `getNextForkByTimestamp returns null when no next fork exists`() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = qbftConsensusConfig)
    val fork3 = ForkSpec(timestampSeconds = 3000L, blockTimeSeconds = 30, configuration = otherConsensusConfig)
    val forks = listOf(fork1, fork2, fork3)

    val schedule = ForksSchedule(expectedChainId, forks)

    // Test getting next fork from last fork timestamp
    assertThat(schedule.getNextForkByTimestamp(3000L)).isNull()

    // Test getting next fork from after last fork
    assertThat(schedule.getNextForkByTimestamp(4000L)).isNull()
  }

  @Test
  fun `getNextForkByTimestamp works with single fork`() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val forks = listOf(fork1)

    val schedule = ForksSchedule(expectedChainId, forks)

    // Test getting next fork from before the only fork
    assertThat(schedule.getNextForkByTimestamp(500L)).isEqualTo(fork1)

    // Test getting next fork from the only fork timestamp
    assertThat(schedule.getNextForkByTimestamp(1000L)).isNull()

    // Test getting next fork from after the only fork
    assertThat(schedule.getNextForkByTimestamp(1500L)).isNull()
  }

  @Test
  fun `getNextForkByTimestamp with edge case timestamps`() {
    val fork1 = ForkSpec(timestampSeconds = 1000L, blockTimeSeconds = 10, configuration = consensusConfig)
    val fork2 = ForkSpec(timestampSeconds = 2000L, blockTimeSeconds = 20, configuration = qbftConsensusConfig)
    val forks = listOf(fork1, fork2)

    val schedule = ForksSchedule(expectedChainId, forks)

    // Test with timestamp exactly one less than first fork
    assertThat(schedule.getNextForkByTimestamp(999L)).isEqualTo(fork1)

    // Test with timestamp exactly one more than first fork
    assertThat(schedule.getNextForkByTimestamp(1001L)).isEqualTo(fork2)

    // Test with timestamp exactly one less than second fork
    assertThat(schedule.getNextForkByTimestamp(1999L)).isEqualTo(fork2)

    // Test with timestamp exactly one more than second fork
    assertThat(schedule.getNextForkByTimestamp(2001L)).isNull()
  }
}
