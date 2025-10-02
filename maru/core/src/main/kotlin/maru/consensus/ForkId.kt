/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import java.time.Clock
import maru.core.Hasher
import maru.database.BeaconChain
import maru.serialization.Serializer
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

data class ForkId(
  val chainId: UInt,
  val forkSpec: ForkSpec,
  val genesisRootHash: ByteArray,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ForkId

    if (chainId != other.chainId) return false
    if (forkSpec != other.forkSpec) return false
    if (!genesisRootHash.contentEquals(other.genesisRootHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = chainId.hashCode()
    result = 31 * result + forkSpec.hashCode()
    result = 31 * result + genesisRootHash.contentHashCode()
    return result
  }
}

class ForkIdHasher(
  val forkIdSerializer: Serializer<ForkId>,
  val hasher: Hasher,
) {
  fun hash(forkId: ForkId): ByteArray = hasher.hash(forkIdSerializer.serialize(forkId)).takeLast(4).toByteArray()
}

interface ForkIdHashManager {
  fun currentHash(): ByteArray

  fun check(otherForkIdHash: ByteArray): Boolean

  fun update(newForkSpec: ForkSpec)
}

class ForkIdHashManagerImpl(
  private val chainId: UInt,
  private val beaconChain: BeaconChain,
  private val forksSchedule: ForksSchedule,
  private val forkIdHasher: ForkIdHasher,
  private val clock: Clock = Clock.systemUTC(),
  private val allowedTimeWindowSeconds: ULong = 20U,
) : ForkIdHashManager {
  data class State(
    var previousForkIdHash: ByteArray?,
    var currentForkIdHash: ByteArray,
    var nextForkIdHash: ByteArray?,
    var currentForkTimestamp: ULong,
    var nextForkTimestamp: ULong,
    var currentBlockTime: UInt,
  )

  private val log: Logger = LogManager.getLogger(this.javaClass)
  private val state: State

  init {
    val timestamp = clock.instant().epochSecond.toULong()
    val previousForkSpec = forksSchedule.getPreviousForkByTimestamp(timestamp)
    val currentForkSpec = forksSchedule.getForkByTimestamp(timestamp)
    val nextForkSpec = forksSchedule.getNextForkByTimestamp(timestamp)

    state =
      State(
        currentForkTimestamp = currentForkSpec.timestampSeconds,
        nextForkTimestamp = nextForkSpec?.timestampSeconds ?: ULong.MAX_VALUE,
        currentBlockTime = forksSchedule.getForkByTimestamp(timestamp).blockTimeSeconds,
        previousForkIdHash = previousForkSpec?.let { getForkIdHashForForkSpec(it) },
        currentForkIdHash = getForkIdHashForForkSpec(currentForkSpec),
        nextForkIdHash = nextForkSpec?.let { getForkIdHashForForkSpec(it) },
      )
  }

  private var _genesisRootHash: ByteArray? = null
  val genesisRootHash: ByteArray
    get() {
      if (_genesisRootHash == null) {
        _genesisRootHash = beaconChain.getBeaconState(0u)?.beaconBlockHeader?.hash
          ?: throw IllegalStateException("Genesis state not found")
      }
      return _genesisRootHash!!
    }

  override fun currentHash(): ByteArray = state.currentForkIdHash

  private fun getForkIdHashForForkSpec(forkSpec: ForkSpec): ByteArray {
    val forkId =
      ForkId(
        chainId = chainId,
        forkSpec = forkSpec,
        genesisRootHash = genesisRootHash,
      )
    return forkIdHasher.hash(forkId)
  }

  override fun check(otherForkIdHash: ByteArray): Boolean {
    if (otherForkIdHash.contentEquals(state.currentForkIdHash)) return true
    val currentTime = clock.instant().epochSecond.toULong()
    // The allowedTimeWindowSeconds allows for time drift, network latency, etc.
    // The poll interval should be the maximum time between the two nodes switching forks (without time drift).
    // Current block time is subtracted from the current fork timestamp, because that is what we do in the fork update logic.
    if (state.previousForkIdHash != null &&
      currentTime <= state.currentForkTimestamp + allowedTimeWindowSeconds &&
      otherForkIdHash.contentEquals(state.previousForkIdHash)
    ) { // this is the case where we have already switched fork
      return true
    }
    if (state.nextForkIdHash != null &&
      currentTime + allowedTimeWindowSeconds >= state.nextForkTimestamp &&
      otherForkIdHash.contentEquals(state.nextForkIdHash)
    ) { // this is the case where we haven't switched fork yet
      return true
    }
    return false
  }

  override fun update(newForkSpec: ForkSpec) {
    log.debug(
      "Updating fork id hash to ${getForkIdHashForForkSpec(newForkSpec).toHexString()} for fork spec=$newForkSpec",
    )
    val newForkIdHash = getForkIdHashForForkSpec(newForkSpec)

    state.previousForkIdHash =
      forksSchedule.getPreviousForkByTimestamp(newForkSpec.timestampSeconds)?.let {
        getForkIdHashForForkSpec(it)
      }

    state.currentForkIdHash = newForkIdHash

    val nextFork = forksSchedule.getNextForkByTimestamp(newForkSpec.timestampSeconds)
    state.nextForkIdHash =
      if (nextFork != null) {
        getForkIdHashForForkSpec(nextFork)
      } else {
        null
      }

    state.currentForkTimestamp = newForkSpec.timestampSeconds
    state.nextForkTimestamp = nextFork?.timestampSeconds ?: ULong.MAX_VALUE

    state.currentBlockTime = newForkSpec.blockTimeSeconds
  }
}
