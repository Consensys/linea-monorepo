/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.core

import kotlin.random.Random
import maru.core.ext.DataGenerators
import maru.database.InMemoryBeaconChain
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class InMemoryBeaconChainTest {
  private lateinit var initialBeaconState: BeaconState
  private lateinit var inMemoryBeaconChain: InMemoryBeaconChain

  @BeforeEach
  fun setUp() {
    initialBeaconState = DataGenerators.randomBeaconState(2UL)
    inMemoryBeaconChain = InMemoryBeaconChain(initialBeaconState)
  }

  @Test
  fun `getLatestBeaconState returns initial state`() {
    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(initialBeaconState)
  }

  @Test
  fun `getBeaconState returns null for unknown block root`() {
    val beaconState = inMemoryBeaconChain.getBeaconState(Random.nextBytes(32))
    assertThat(beaconState).isNull()
  }

  @Test
  fun `getSealedBeaconBlock returns null for unknown block root`() {
    val sealedBeaconBlock = inMemoryBeaconChain.getSealedBeaconBlock(Random.nextBytes(32))
    assertThat(sealedBeaconBlock).isNull()
  }

  @Test
  fun `getSealedBeaconBlock returns null for unknown block number`() {
    val sealedBeaconBlock = inMemoryBeaconChain.getSealedBeaconBlock(100uL)
    assertThat(sealedBeaconBlock).isNull()
  }

  @Test
  fun `newUpdater can put and commit beacon state`() {
    val newBeaconState = DataGenerators.randomBeaconState(2UL)
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putBeaconState(newBeaconState).commit()

    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(newBeaconState)

    val retrievedBeaconState = inMemoryBeaconChain.getBeaconState(newBeaconState.latestBeaconBlockHeader.hash)
    assertThat(retrievedBeaconState).isEqualTo(newBeaconState)
  }

  @Test
  fun `newUpdater can put and commit sealed beacon block`() {
    val sealedBeaconBlock = DataGenerators.randomSealedBeaconBlock(3UL)
    val beaconBlockRoot = sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putSealedBeaconBlock(sealedBeaconBlock).commit()

    val retrievedSealedBeaconBlockByBlockRoot = inMemoryBeaconChain.getSealedBeaconBlock(beaconBlockRoot)
    assertThat(retrievedSealedBeaconBlockByBlockRoot).isEqualTo(sealedBeaconBlock)

    val retrievedSealedBeaconBlockByBlockNumber =
      inMemoryBeaconChain
        .getSealedBeaconBlock(sealedBeaconBlock.beaconBlock.beaconBlockHeader.number)
    assertThat(retrievedSealedBeaconBlockByBlockNumber).isEqualTo(sealedBeaconBlock)
  }

  @Test
  fun `newUpdater can rollback changes`() {
    val newBeaconState = DataGenerators.randomBeaconState(4UL)
    val sealedBeaconBlock = DataGenerators.randomSealedBeaconBlock(5UL)
    val beaconBlockRoot = sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putBeaconState(newBeaconState)
    updater.putSealedBeaconBlock(sealedBeaconBlock)
    updater.rollback()

    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(initialBeaconState)

    val retrievedBeaconState = inMemoryBeaconChain.getBeaconState(newBeaconState.latestBeaconBlockHeader.hash)
    assertThat(retrievedBeaconState).isNull()

    val retrievedSealedBeaconBlockByBlockRoot = inMemoryBeaconChain.getSealedBeaconBlock(beaconBlockRoot)
    assertThat(retrievedSealedBeaconBlockByBlockRoot).isNull()

    val retrievedSealedBeaconBlockByBlockNumber =
      inMemoryBeaconChain
        .getSealedBeaconBlock(sealedBeaconBlock.beaconBlock.beaconBlockHeader.number)
    assertThat(retrievedSealedBeaconBlockByBlockNumber).isNull()
  }

  @Test
  fun `uncommited changes are not visible by InMemoryBeaconChain`() {
    val newBeaconState = DataGenerators.randomBeaconState(6UL)
    val newBeaconBlock = DataGenerators.randomSealedBeaconBlock(7UL)
    val inflightBeaconBlockRoot = newBeaconBlock.beaconBlock.beaconBlockHeader.hash
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putBeaconState(newBeaconState)
    updater.putSealedBeaconBlock(newBeaconBlock)

    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(initialBeaconState)

    val retrievedBeaconState = inMemoryBeaconChain.getBeaconState(newBeaconState.latestBeaconBlockHeader.hash)
    assertThat(retrievedBeaconState).isNull()

    val retrievedSealedBeaconBlockByBlockRoot = inMemoryBeaconChain.getSealedBeaconBlock(inflightBeaconBlockRoot)
    assertThat(retrievedSealedBeaconBlockByBlockRoot).isNull()

    val retrievedSealedBeaconBlockByBlockNumber =
      inMemoryBeaconChain
        .getSealedBeaconBlock(newBeaconBlock.beaconBlock.beaconBlockHeader.number)
    assertThat(retrievedSealedBeaconBlockByBlockNumber).isNull()
  }

  @Test
  fun `initial state can be found by hash`() {
    val initialBeaconStateByHash = inMemoryBeaconChain.getBeaconState(initialBeaconState.latestBeaconBlockHeader.hash)
    assertThat(initialBeaconStateByHash).isEqualTo(initialBeaconState)
  }

  @Test
  fun `initial state can be found by number`() {
    val initialBeaconStateByNumber =
      inMemoryBeaconChain.getBeaconState(initialBeaconState.latestBeaconBlockHeader.number)
    assertThat(initialBeaconStateByNumber).isEqualTo(initialBeaconState)
  }
}
