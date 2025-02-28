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

import java.util.NavigableSet
import java.util.TreeSet

interface ConsensusConfig

data class ForkSpec(
  val blockNumber: ULong,
  val configuration: ConsensusConfig,
)

class ForksSchedule(
  forks: Collection<ForkSpec>,
) {
  private val forks: NavigableSet<ForkSpec> =
    run {
      val newForks =
        TreeSet(
          Comparator.comparing(ForkSpec::blockNumber).reversed(),
        )
      newForks.addAll(forks)
      newForks
    }

  fun getForkByNumber(blockNumber: ULong): ConsensusConfig {
    for (f in forks) {
      if (blockNumber >= f.blockNumber) {
        return f.configuration
      }
    }

    return forks.first().configuration
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ForksSchedule

    return forks == other.forks
  }

  override fun hashCode(): Int = forks.hashCode()
}
