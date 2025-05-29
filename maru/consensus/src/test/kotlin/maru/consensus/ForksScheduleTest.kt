/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.consensus

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows

class ForksScheduleTest {
  private val consensusConfig = object : ConsensusConfig {}
  private val expectedChainId = 1337u

  @Test
  fun `test getForkByTimestamp returns correct fork`() {
    val fork1 = ForkSpec(1L, 1, consensusConfig)
    val fork2 = ForkSpec(2L, 2, consensusConfig)
    val fork3 = ForkSpec(3L, 3, consensusConfig)
    val forks = listOf(fork1, fork2, fork3)

    val schedule = ForksSchedule(expectedChainId, forks)

    assertThat(schedule.getForkByTimestamp(1L)).isEqualTo(fork1)
    assertThat(schedule.getForkByTimestamp(2L)).isEqualTo(fork2)
    assertThat(schedule.getForkByTimestamp(3L)).isEqualTo(fork3)
  }

  @Test
  fun `getForkByTimestamp throws if timestamp is before all forks`() {
    val fork1 = ForkSpec(1000L, 10, consensusConfig)
    val fork2 = ForkSpec(2000L, 20, consensusConfig)
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
        ForkSpec(1000L, 0, consensusConfig)
      }
    assertThat(exception).hasMessage("blockTimeSeconds must be greater or equal to 1 second")
  }

  @Test
  fun equality() {
    val fork1 = ForkSpec(1000L, 10, consensusConfig)
    val fork2 = ForkSpec(2000L, 20, consensusConfig)
    val forks1 = listOf(fork1, fork2)
    val forks2 = listOf(fork1, fork2)

    val schedule1 = ForksSchedule(expectedChainId, forks1)
    val schedule2 = ForksSchedule(expectedChainId, forks2)

    assertThat(schedule1).isEqualTo(schedule2)
    assertThat(schedule1.hashCode()).isEqualTo(schedule2.hashCode())
  }

  @Test
  fun inequality() {
    val fork1 = ForkSpec(1000L, 10, consensusConfig)
    val fork2 = ForkSpec(2000L, 20, consensusConfig)
    val fork3 = ForkSpec(3000L, 30, consensusConfig)
    val forks1 = listOf(fork1, fork2)
    val forks2 = listOf(fork1, fork3)

    val schedule1 = ForksSchedule(expectedChainId, forks1)
    val schedule2 = ForksSchedule(expectedChainId, forks2)

    assertThat(schedule1).isNotEqualTo(schedule2)
    assertThat(schedule1.hashCode()).isNotEqualTo(schedule2.hashCode())
  }
}
