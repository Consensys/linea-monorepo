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
  fun `newUpdater can put and commit beacon state`() {
    val newBeaconState = DataGenerators.randomBeaconState(2UL)
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putBeaconState(newBeaconState).commit()

    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(newBeaconState)

    val retrievedBeaconState = inMemoryBeaconChain.getBeaconState(newBeaconState.latestBeaconBlockRoot)
    assertThat(retrievedBeaconState).isEqualTo(newBeaconState)
  }

  @Test
  fun `newUpdater can put and commit sealed beacon block`() {
    val sealedBeaconBlock = DataGenerators.randomSealedBeaconBlock(3UL)
    val beaconBlockRoot = sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putSealedBeaconBlock(sealedBeaconBlock, beaconBlockRoot).commit()

    val retrievedSealedBeaconBlock = inMemoryBeaconChain.getSealedBeaconBlock(beaconBlockRoot)
    assertThat(retrievedSealedBeaconBlock).isEqualTo(sealedBeaconBlock)
  }

  @Test
  fun `newUpdater can rollback changes`() {
    val newBeaconState = DataGenerators.randomBeaconState(4UL)
    val sealedBeaconBlock = DataGenerators.randomSealedBeaconBlock(5UL)
    val beaconBlockRoot = sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putBeaconState(newBeaconState)
    updater.putSealedBeaconBlock(sealedBeaconBlock, beaconBlockRoot)
    updater.rollback()

    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(initialBeaconState)

    val retrievedBeaconState = inMemoryBeaconChain.getBeaconState(newBeaconState.latestBeaconBlockRoot)
    assertThat(retrievedBeaconState).isNull()

    val retrievedSealedBeaconBlock = inMemoryBeaconChain.getSealedBeaconBlock(beaconBlockRoot)
    assertThat(retrievedSealedBeaconBlock).isNull()
  }

  @Test
  fun `uncommited changes are not visible by InMemoryBeaconChain`() {
    val newBeaconState = DataGenerators.randomBeaconState(6UL)
    val newBeaconBlock = DataGenerators.randomSealedBeaconBlock(7UL)
    val inflightBeaconBlockRoot = newBeaconBlock.beaconBlock.beaconBlockHeader.hash
    val updater = inMemoryBeaconChain.newUpdater()
    updater.putBeaconState(newBeaconState)
    updater.putSealedBeaconBlock(newBeaconBlock, inflightBeaconBlockRoot)

    val latestBeaconState = inMemoryBeaconChain.getLatestBeaconState()
    assertThat(latestBeaconState).isEqualTo(initialBeaconState)

    val retrievedBeaconState = inMemoryBeaconChain.getBeaconState(newBeaconState.latestBeaconBlockRoot)
    assertThat(retrievedBeaconState).isNull()

    val retrievedSealedBeaconBlock = inMemoryBeaconChain.getSealedBeaconBlock(inflightBeaconBlockRoot)
    assertThat(retrievedSealedBeaconBlock).isNull()
  }
}
