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

import java.math.BigInteger
import java.util.Collections
import maru.consensus.ValidatorProvider
import maru.consensus.qbft.adapters.QbftBlockAdapter
import maru.consensus.qbft.adapters.QbftBlockHeaderAdapter
import maru.consensus.qbft.adapters.toBeaconBlock
import maru.consensus.qbft.adapters.toBeaconBlockHeader
import maru.consensus.qbft.adapters.toSealedBeaconBlock
import maru.core.BeaconState
import maru.core.HashUtil
import maru.core.Seal
import maru.core.Validator
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import maru.serialization.rlp.bodyRoot
import maru.serialization.rlp.headerHash
import maru.serialization.rlp.stateRoot
import org.assertj.core.api.Assertions
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.blockcreation.ProposerSelector
import org.hyperledger.besu.crypto.SECPSignature
import org.hyperledger.besu.datatypes.Address
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.async.SafeFuture.completedFuture

class QbftBlockCreatorTest {
  private val executionLayerManager = Mockito.mock(ExecutionLayerManager::class.java)
  private val proposerSelector = Mockito.mock(ProposerSelector::class.java)
  private val validatorProvider = Mockito.mock(ValidatorProvider::class.java)
  private val beaconChain = Mockito.mock(BeaconChain::class.java)

  @Test
  fun `can create block`() {
    val parentBlock = DataGenerators.randomSealedBeaconBlock(10U)
    val parentHeader = QbftBlockHeaderAdapter(parentBlock.beaconBlock.beaconBlockHeader)
    val executionPayload = DataGenerators.randomExecutionPayload()
    whenever(beaconChain.getSealedBeaconBlock(parentBlock.beaconBlock.beaconBlockHeader.hash())).thenReturn(parentBlock)
    whenever(executionLayerManager.finishBlockBuilding()).thenReturn(completedFuture(executionPayload))
    whenever(proposerSelector.selectProposerForRound(ConsensusRoundIdentifier(11L, 1))).thenReturn(Address.ZERO)
    whenever(
      validatorProvider.getValidatorsAfterBlock(10U),
    ).thenReturn(completedFuture(DataGenerators.randomValidators()))

    val blockCreator = QbftBlockCreator(executionLayerManager, proposerSelector, validatorProvider, beaconChain, 1)
    val createdBlock = blockCreator.createBlock(1000L, parentHeader)
    val createBeaconBlock = createdBlock.toBeaconBlock()

    // block header fields
    val blockHeader = createBeaconBlock.beaconBlockHeader
    assertThat(blockHeader.number).isEqualTo(11UL)
    assertThat(blockHeader.round).isEqualTo(1UL)
    assertThat(blockHeader.timestamp).isEqualTo(1000UL)
    assertThat(blockHeader.proposer).isEqualTo(Validator(Address.ZERO.toArray()))

    // block header roots
    val stateRoot =
      HashUtil.stateRoot(
        BeaconState(
          createBeaconBlock.beaconBlockHeader.copy(stateRoot = ByteArray(32)),
          HashUtil.bodyRoot(createBeaconBlock.beaconBlockBody),
          Collections.emptySet(),
        ),
      )
    Assertions
      .assertThat(
        blockHeader.bodyRoot,
      ).isEqualTo(
        HashUtil.bodyRoot(createBeaconBlock.beaconBlockBody),
      )
    Assertions.assertThat(blockHeader.stateRoot).isEqualTo(stateRoot)
    assertThat(blockHeader.parentRoot).isEqualTo(parentHeader.toBeaconBlockHeader().hash())
    assertThat(
      createBeaconBlock.beaconBlockHeader.hash(),
    ).isEqualTo(HashUtil.headerHash(createBeaconBlock.beaconBlockHeader))

    // block body fields
    val blockBody = createBeaconBlock.beaconBlockBody
    assertThat(
      blockBody.prevCommitSeals,
    ).isEqualTo(
      parentBlock.commitSeals,
    )
    assertThat(blockBody.executionPayload).isEqualTo(executionPayload)
  }

  @Test
  fun `fails to create block if execution payload not available`() {
    val parentBlock = DataGenerators.randomBeaconBlock(10U)
    val parentHeader = QbftBlockHeaderAdapter(parentBlock.beaconBlockHeader)

    whenever(
      executionLayerManager.finishBlockBuilding(),
    ).thenReturn(SafeFuture.failedFuture(IllegalStateException("Execution payload not available")))

    val blockCreator = QbftBlockCreator(executionLayerManager, proposerSelector, validatorProvider, beaconChain, 1)
    Assertions
      .assertThatThrownBy({
        blockCreator.createBlock(1000L, parentHeader)
      })
      .isInstanceOf(
        IllegalStateException::class.java,
      ).hasMessage("Execution payload unavailable, unable to create block")
  }

  @Test
  fun `fails to create block if parent beacon block not available`() {
    val parentBlock = DataGenerators.randomBeaconBlock(10U)
    val parentHeader = QbftBlockHeaderAdapter(parentBlock.beaconBlockHeader)
    val executionPayload = DataGenerators.randomExecutionPayload()

    whenever(executionLayerManager.finishBlockBuilding()).thenReturn(completedFuture(executionPayload))
    whenever(beaconChain.getSealedBeaconBlock(parentBlock.beaconBlockHeader.hash())).thenReturn(null)
    whenever(proposerSelector.selectProposerForRound(ConsensusRoundIdentifier(11L, 1))).thenReturn(Address.ZERO)

    val blockCreator = QbftBlockCreator(executionLayerManager, proposerSelector, validatorProvider, beaconChain, 1)
    Assertions
      .assertThatThrownBy({
        blockCreator.createBlock(1000L, parentHeader)
      })
      .isInstanceOf(
        IllegalStateException::class.java,
      ).hasMessage("Parent beacon block unavailable, unable to create block")
  }

  @Test
  fun `can create sealed block`() {
    val block = QbftBlockAdapter(DataGenerators.randomBeaconBlock(10U))
    val beaconBlock = block.toBeaconBlock()
    val seals = listOf(SECPSignature.create(BigInteger.ONE, BigInteger.TWO, 0x00, BigInteger.valueOf(4)))
    val round = 1

    val blockCreator = QbftBlockCreator(executionLayerManager, proposerSelector, validatorProvider, beaconChain, 1)
    val createSealedBlock = blockCreator.createSealedBlock(block, round, seals)
    val createdSealedBeaconBlock = createSealedBlock.toSealedBeaconBlock()

    // block header fields
    val createdSealedBlockHeader = createdSealedBeaconBlock.beaconBlock.beaconBlockHeader
    assertThat(createdSealedBlockHeader.number).isEqualTo(block.header.number.toULong())
    assertThat(createdSealedBlockHeader.round).isEqualTo(round.toULong()) // round number is updated
    assertThat(createdSealedBlockHeader.timestamp).isEqualTo(block.header.timestamp.toULong())
    assertThat(createdSealedBlockHeader.proposer).isEqualTo(block.toBeaconBlock().beaconBlockHeader.proposer)

    // block header roots
    Assertions
      .assertThat(
        createdSealedBlockHeader.bodyRoot,
      ).isEqualTo(
        beaconBlock.beaconBlockHeader.bodyRoot,
      )
    Assertions.assertThat(createdSealedBlockHeader.stateRoot).isEqualTo(beaconBlock.beaconBlockHeader.stateRoot)
    assertThat(createdSealedBlockHeader.parentRoot).isEqualTo(beaconBlock.beaconBlockHeader.parentRoot)
    assertThat(createdSealedBlockHeader.hash()).isEqualTo(HashUtil.headerHash(createdSealedBlockHeader))

    // block body fields
    val sealedBlockBody = createdSealedBeaconBlock.beaconBlock.beaconBlockBody
    assertThat(
      sealedBlockBody.prevCommitSeals,
    ).isEqualTo(
      beaconBlock.beaconBlockBody.prevCommitSeals,
    )
    val beaconSeals = seals.map { Seal(it.encodedBytes().toArrayUnsafe()) }
    assertThat(createdSealedBeaconBlock.commitSeals).isEqualTo(beaconSeals)
    assertThat(sealedBlockBody.executionPayload).isEqualTo(beaconBlock.beaconBlockBody.executionPayload)
  }
}
