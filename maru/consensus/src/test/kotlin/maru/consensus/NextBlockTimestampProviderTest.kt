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

import java.time.Clock
import java.time.Instant
import java.time.ZoneId
import kotlin.test.Test
import maru.consensus.qbft.QbftConsensusConfig
import org.assertj.core.api.Assertions.assertThat

class NextBlockTimestampProviderTest {
  private val chainId = 1337u
  private val forksSchedule =
    ForksSchedule(
      chainId,
      listOf(
        ForkSpec(0, 1, QbftConsensusConfig(validatorSet = emptySet(), ElFork.Prague)),
        ForkSpec(10, 2, QbftConsensusConfig(validatorSet = emptySet(), ElFork.Prague)),
      ),
    )
  private val baseLastBlockTimestamp = 9L

  private fun createCLockForTimestamp(timestamp: Long): Clock =
    Clock.fixed(Instant.ofEpochMilli(timestamp), ZoneId.of("UTC"))

  @Test
  fun `nextBlockTimestampProvider targets next planned block timestamp`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(9999L),
        forksSchedule,
      )

    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(baseLastBlockTimestamp)

    assertThat(result).isEqualTo(10L)
  }

  @Test
  fun `if current time is overdue it targets next integer second`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(11123),
        forksSchedule,
      )

    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(baseLastBlockTimestamp)

    assertThat(result).isEqualTo(12L)
  }

  @Test
  fun `nextBlockTimestampProvider takes into account forks schedule`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(11123),
        forksSchedule,
      )
    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(10L)

    assertThat(result).isEqualTo(12L)
  }
}
