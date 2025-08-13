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
import kotlin.reflect.KClass
import org.apache.logging.log4j.LogManager

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
  private val log = LogManager.getLogger(this.javaClass)
  private val forks: NavigableSet<ForkSpec> =
    run {
      val newForks =
        TreeSet(
          Comparator
            .comparing(ForkSpec::timestampSeconds)
            .reversed(),
        )
      newForks.addAll(forks)
      require(newForks.size == forks.size) { "Fork timestamps must be unique" }
      newForks
    }

  fun getForkByTimestamp(timestamp: Long): ForkSpec =
    forks.firstOrNull { timestamp >= it.timestampSeconds } ?: throw IllegalArgumentException(
      "No fork found for $timestamp, first known fork is at ${forks.last.timestampSeconds}",
    )

  fun <T : ConsensusConfig> getForkByConfigType(configClass: KClass<T>): ForkSpec {
    // Uses findLast since the list is reversed to get the first matching fork
    val fork = forks.findLast { it.configuration::class == configClass }
    return fork ?: throw IllegalArgumentException(
      "No fork found for config type ${configClass.simpleName}",
    )
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ForksSchedule

    return forks == other.forks
  }

  override fun hashCode(): Int = forks.hashCode()

  override fun toString(): String = "ForksSchedule(chainId=$chainId, forks=$forks)"
}
