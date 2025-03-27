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
import kotlin.math.ceil
import kotlin.math.max
import kotlin.time.Duration
import maru.executionlayer.manager.BlockMetadata

fun interface NextBlockTimestampProvider {
  fun nextTargetBlockUnixTimestamp(lastBlockMetadata: BlockMetadata): Long
}

class NextBlockTimestampProviderImpl(
  private val clock: Clock,
  private val forksSchedule: ForksSchedule,
  private val minTimeTillNextBlock: Duration,
) : NextBlockTimestampProvider {
  override fun nextTargetBlockUnixTimestamp(lastBlockMetadata: BlockMetadata): Long {
    val currentBlockTime = forksSchedule.getForkByTimestamp(lastBlockMetadata.unixTimestampSeconds).blockTimeSeconds
    val lastBlockTimestamp = lastBlockMetadata.unixTimestampSeconds

    val nextIntegerSecond = ceil((clock.millis() + minTimeTillNextBlock.inWholeMilliseconds) / 1000.0).toLong()
    return max(lastBlockTimestamp + currentBlockTime, nextIntegerSecond)
  }
}
