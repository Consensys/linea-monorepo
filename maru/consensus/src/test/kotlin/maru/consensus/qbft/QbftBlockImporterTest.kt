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
package maru.consensus.qbft

import com.github.michaelbull.result.Ok
import kotlin.test.assertFalse
import kotlin.test.assertTrue
import maru.consensus.qbft.adapters.QbftSealedBlockAdapter
import maru.consensus.state.StateTransition
import maru.core.BeaconState
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import maru.database.InMemoryBeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito.mock
import org.mockito.Mockito.reset
import org.mockito.kotlin.any
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture

class QbftBlockImporterTest {
  private var executionLayerManager: ExecutionLayerManager = mock()

  private val stateTransition: StateTransition = mock()

  private lateinit var beaconChain: BeaconChain
  private var beaconBlockImporterResponse =
    SafeFuture.completedFuture(DataGenerators.randomValidForkChoiceUpdatedResult())

  private lateinit var qbftBlockImporter: QbftBlockImportCoordinator
  private lateinit var initialBeaconState: BeaconState

  @BeforeEach
  fun setUp() {
    initialBeaconState = DataGenerators.randomBeaconState(2UL)
    beaconChain = InMemoryBeaconChain(initialBeaconState)
    qbftBlockImporter =
      QbftBlockImportCoordinator(
        beaconChain = beaconChain,
        stateTransition = stateTransition,
        beaconBlockImporter = { beaconBlockImporterResponse },
      )
  }

  @AfterEach
  fun tearDown() {
    reset(executionLayerManager)
    reset(stateTransition)
  }

  @Test
  fun `importBlock returns true on successful import`() {
    val qbftBlock = QbftSealedBlockAdapter(DataGenerators.randomSealedBeaconBlock(2UL))
    val beaconState = DataGenerators.randomBeaconState(2UL)

    whenever(stateTransition.processBlock(any(), any())).thenReturn(SafeFuture.completedFuture(Ok(beaconState)))
    whenever(executionLayerManager.setHead(any(), any(), any())).thenReturn(
      SafeFuture.completedFuture(
        DataGenerators
          .randomValidForkChoiceUpdatedResult(),
      ),
    )

    val result = qbftBlockImporter.importBlock(qbftBlock)
    assertTrue(result)
    assertThat(beaconChain.getLatestBeaconState()).isEqualTo(beaconState)
  }

  @Test
  fun `importBlock rolls the DB update back on state transition failure`() {
    val qbftBlock = QbftSealedBlockAdapter(DataGenerators.randomSealedBeaconBlock(2UL))
    whenever(stateTransition.processBlock(any(), any())).thenThrow(RuntimeException("Test exception"))
    val stateBeforeTransition = beaconChain.getLatestBeaconState()

    val result = qbftBlockImporter.importBlock(qbftBlock)

    assertFalse(result)
    val stateAfterTransition = beaconChain.getLatestBeaconState()
    assertThat(stateBeforeTransition).isEqualTo(stateAfterTransition)
  }

  @Test
  fun `importBlock rolls the DB update back on block import`() {
    val qbftBlock = QbftSealedBlockAdapter(DataGenerators.randomSealedBeaconBlock(2UL))
    beaconBlockImporterResponse = SafeFuture.failedFuture(RuntimeException("Test exception"))
    val stateBeforeTransition = beaconChain.getLatestBeaconState()

    val result = qbftBlockImporter.importBlock(qbftBlock)

    assertFalse(result)
    val stateAfterTransition = beaconChain.getLatestBeaconState()
    assertThat(stateBeforeTransition).isEqualTo(stateAfterTransition)
  }
}
