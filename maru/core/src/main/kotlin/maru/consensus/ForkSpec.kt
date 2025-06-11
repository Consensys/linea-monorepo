/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import java.util.NavigableSet
import java.util.TreeSet

data class ForkSpec(
  val timestampSeconds: Long,
  val blockTimeSeconds: Int,
  val configuration: ConsensusConfig,
) {
  init {
    require(blockTimeSeconds > 0) { "blockTimeSeconds must be greater or equal to 1 second" }
  }
}

class ForksSchedule(
  val chainId: UInt,
  forks: Collection<ForkSpec>,
) {
  private val forks: NavigableSet<ForkSpec> =
    run {
      val newForks =
        TreeSet(
          Comparator.comparing(ForkSpec::timestampSeconds).reversed(),
        )
      newForks.addAll(forks)
      newForks
    }

  fun getForkByTimestamp(timestamp: Long): ForkSpec {
    for (f in forks) {
      if (timestamp >= f.timestampSeconds) {
        return f
      }
    }

    throw IllegalArgumentException(
      "No fork found for $timestamp, first known fork is at ${forks.last.timestampSeconds}",
    )
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ForksSchedule

    return forks == other.forks
  }

  override fun hashCode(): Int = forks.hashCode()
}
