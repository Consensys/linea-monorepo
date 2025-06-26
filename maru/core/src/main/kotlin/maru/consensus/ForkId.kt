/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import maru.core.Hasher
import maru.database.BeaconChain
import maru.serialization.Serializer

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

class ForkIdHashProvider(
  private val chainId: UInt,
  private val beaconChain: BeaconChain,
  private val forksSchedule: ForksSchedule,
  private val forkIdHasher: ForkIdHasher,
) {
  fun currentForkIdHash(): ByteArray {
    val forkId =
      ForkId(
        chainId = chainId,
        forkSpec =
          forksSchedule.getForkByTimestamp(
            beaconChain
              .getLatestBeaconState()
              .latestBeaconBlockHeader.timestamp
              .toLong(),
          ),
        genesisRootHash =
          beaconChain.getBeaconState(0u)?.latestBeaconBlockHeader?.hash
            ?: throw IllegalStateException("Genesis state not found"),
      )
    return forkIdHasher.hash(forkId)
  }
}
