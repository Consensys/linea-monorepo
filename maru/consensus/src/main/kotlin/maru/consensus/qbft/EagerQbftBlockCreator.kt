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

import kotlin.time.Duration
import maru.consensus.qbft.adapters.toBeaconBlockHeader
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlockHeader
import maru.core.Validator
import maru.executionlayer.manager.ExecutionLayerManager
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlock
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockCreator
import org.hyperledger.besu.consensus.qbft.core.types.QbftBlockHeader
import org.hyperledger.besu.crypto.SECPSignature

/**
 * Responsible for QBFT block creation. As opposed to DelayedQbftBlockCreator, Eager one will send the FCU request to
 * the execution client to start the block building process and will wait for the required time until a block is built,
 * blocking the thread
 */
class EagerQbftBlockCreator(
  private val manager: ExecutionLayerManager,
  private val delegate: QbftBlockCreator,
  private val finalizationStateProvider: (BeaconBlockHeader) -> FinalizationState,
  private val blockBuilderIdentity: Validator,
  private val config: Config,
) : QbftBlockCreator {
  data class Config(
    val blockBuildingDuration: Duration,
  )

  override fun createBlock(
    headerTimeStampSeconds: Long,
    parentHeader: QbftBlockHeader,
  ): QbftBlock {
    val beaconBlockHeader = parentHeader.toBeaconBlockHeader()
    val finalizedState = finalizationStateProvider(beaconBlockHeader)
    manager
      .setHeadAndStartBlockBuilding(
        headHash = manager.latestBlockMetadata().blockHash,
        safeHash = finalizedState.safeBlockHash,
        finalizedHash = finalizedState.finalizedBlockHash,
        nextBlockTimestamp = headerTimeStampSeconds,
        feeRecipient = blockBuilderIdentity.address,
      ).get()
    Thread.sleep(config.blockBuildingDuration.inWholeMilliseconds)
    return delegate.createBlock(headerTimeStampSeconds, parentHeader)
  }

  override fun createSealedBlock(
    block: QbftBlock,
    roundNumber: Int,
    commitSeals: MutableCollection<SECPSignature>,
  ): QbftBlock = DelayedQbftBlockCreator.createSealedBlock(block, roundNumber, commitSeals)
}
