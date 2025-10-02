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
import java.time.Instant
import java.time.ZoneOffset
import maru.core.Hasher
import maru.core.ext.DataGenerators
import maru.database.InMemoryBeaconChain
import maru.serialization.Serializer
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.kotlin.whenever

class ForkIdHashManagerImplTest {
  private val initialBeaconState =
    DataGenerators.randomBeaconState(number = 0UL)
  private val beaconChain = InMemoryBeaconChain(initialBeaconState)

  private val chainId: UInt = 1u

  private val forkSpec1 = ForkSpec(10UL, 1U, mock<ConsensusConfig>())
  private val forkSpec2 = ForkSpec(20UL, 1U, mock<ConsensusConfig>())
  private val forkSpec3 = ForkSpec(30UL, 1U, mock<ConsensusConfig>())
  private var forksSchedule = ForksSchedule(1U, listOf(forkSpec1, forkSpec2, forkSpec3))
  private val forkIdHasher by lazy { ForkIdHasher(serializer, hasher) }
  private val genesisHash = initialBeaconState.beaconBlockHeader.hash
  private var clock = Clock.fixed(Instant.ofEpochSecond(20), ZoneOffset.UTC)

  private val forkId1 = ForkId(chainId, forkSpec1, genesisHash)
  private val forkId2 = ForkId(chainId, forkSpec2, genesisHash)
  private val forkId3 = ForkId(chainId, forkSpec3, genesisHash)

  private val serializedFork1 = ByteArray(32) { 1 }
  private val serializedFork2 = ByteArray(32) { 2 }
  private val serializedFork3 = ByteArray(32) { 3 }

  private val serializer =
    Serializer<ForkId> { value ->
      when (value) {
        forkId1 -> serializedFork1
        forkId2 -> serializedFork2
        forkId3 -> serializedFork3
        else -> throw IllegalArgumentException("Unknown ForkId")
      }
    }

  private val hashFork1 = ByteArray(4) { 1 }
  private val hashFork2 = ByteArray(4) { 2 }
  private val hashFork3 = ByteArray(4) { 3 }
  private val hashUnknownFork = ByteArray(4) { 9 }

  private val hasher =
    Hasher { value ->
      when (value) {
        serializedFork1 -> hashFork1
        serializedFork2 -> hashFork2
        serializedFork3 -> hashFork3
        else -> throw IllegalArgumentException("Unknown serialized value")
      }
    }

  @Test
  fun `currentHash returns correct hash for current fork`() {
    val manager = ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock)
    val hash = manager.currentHash()
    assertThat(hash).isEqualTo(hashFork2)
  }

  @Test
  fun `check returns true for current fork id hash`() {
    val manager = ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock)
    val currentHash = manager.currentHash()
    assertThat(manager.check(currentHash)).isTrue()
  }

  @Test
  fun `check returns false for unknown fork id hash`() {
    val manager = ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock)
    assertThat(manager.check(hashUnknownFork)).isFalse
  }

  @Test
  fun `update changes current fork id hash to next fork`() {
    val manager = ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock)

    manager.update(forkSpec3)
    assertThat(manager.currentHash()).isEqualTo(hashFork3)
  }

  @Test
  fun `check returns true for previous fork id hash within allowed time window`() {
    val manager = ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock)

    // Get previous fork id hash
    val previousForkId = ForkId(chainId, forkSpec1, genesisHash)
    val previousForkIdHash = forkIdHasher.hash(previousForkId)

    // Should be within allowed window, as clock is constant and at timestamp for the current fork
    assertThat(manager.check(previousForkIdHash)).isTrue
  }

  @Test
  fun `check returns true for next fork id hash within allowed time window`() {
    val clock = mock<Clock>()
    whenever(clock.instant()).thenReturn(Instant.ofEpochSecond(20), Instant.ofEpochSecond(29))
    val manager = ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock)

    // Get previous fork id hash
    val nextForkId = ForkId(chainId, forkSpec3, genesisHash)
    val nextForkIdHash = forkIdHasher.hash(nextForkId)

    // Should be within allowed window
    assertThat(manager.check(nextForkIdHash)).isTrue
  }

  @Test
  fun `check returns false for next fork id hash not within allowed time window`() {
    val clock = mock<Clock>()
    whenever(clock.instant()).thenReturn(Instant.ofEpochSecond(20), Instant.ofEpochSecond(21))
    val manager =
      ForkIdHashManagerImpl(chainId, beaconChain, forksSchedule, forkIdHasher, clock, allowedTimeWindowSeconds = 0UL)

    // Get previous fork id hash
    val nextForkId = ForkId(chainId, forkSpec3, genesisHash)
    val nextForkIdHash = forkIdHasher.hash(nextForkId)

    // Should be outside the allowed window
    assertThat(manager.check(nextForkIdHash)).isFalse
  }
}
