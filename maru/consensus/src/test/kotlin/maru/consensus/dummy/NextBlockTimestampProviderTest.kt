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
package maru.consensus.dummy

import java.time.Clock
import java.time.Instant
import java.time.ZoneId
import kotlin.test.Test
import kotlin.time.Duration.Companion.milliseconds
import maru.consensus.ElFork
import maru.consensus.ForkSpec
import maru.consensus.ForksSchedule
import maru.executionlayer.manager.BlockMetadata
import org.assertj.core.api.Assertions.assertThat

class NextBlockTimestampProviderTest {
  private val forksSchedule =
    ForksSchedule(
      listOf(
        ForkSpec(0, 1, DummyConsensusConfig(ByteArray(20), ElFork.Prague)),
        ForkSpec(10, 2, DummyConsensusConfig(ByteArray(20), ElFork.Prague)),
      ),
    )
  private val baseLastBlockMetadata = BlockMetadata(1UL, ByteArray(32), 9)

  private fun createCLockForTimestamp(timestamp: Long): Clock =
    Clock.fixed(Instant.ofEpochMilli(timestamp), ZoneId.of("UTC"))

  @Test
  fun `nextBlockTimestampProvider targets next planned block timestamp`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(9999L),
        forksSchedule,
        minTimeTillNextBlock = 0.milliseconds,
      )

    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(baseLastBlockMetadata)

    assertThat(result).isEqualTo(10L)
  }

  @Test
  fun `if current time till next block is too short, it returns next integer second`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(9901L),
        forksSchedule,
        minTimeTillNextBlock = 100.milliseconds,
      )

    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(baseLastBlockMetadata)

    assertThat(result).isEqualTo(11L)
  }

  @Test
  fun `if current time is overdue it targets next integer second`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(11123),
        forksSchedule,
        minTimeTillNextBlock = 100.milliseconds,
      )

    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(baseLastBlockMetadata)

    assertThat(result).isEqualTo(12L)
  }

  @Test
  fun `nextBlockTimestampProvider takes into account forks schedule`() {
    val nextBlockTimestampProvider =
      NextBlockTimestampProviderImpl(
        createCLockForTimestamp(11123),
        forksSchedule,
        minTimeTillNextBlock = 100.milliseconds,
      )

    val lastBlockMetadata = baseLastBlockMetadata.copy(unixTimestampSeconds = 10L)
    val result = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(lastBlockMetadata)

    assertThat(result).isEqualTo(12L)
  }
}
