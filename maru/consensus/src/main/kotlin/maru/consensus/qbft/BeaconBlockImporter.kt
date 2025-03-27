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

import maru.consensus.state.FinalizationState
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.Validator
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface BeaconBlockImporter {
  fun importBlock(beaconBlock: BeaconBlock): SafeFuture<ForkChoiceUpdatedResult>
}

class BeaconBlockImporterImpl(
  private val executionLayerManager: ExecutionLayerManager,
  private val finalizationStateProvider: (BeaconBlockHeader) -> FinalizationState,
  private val nextBlockTimestampProvider: (ConsensusRoundIdentifier) -> Long,
  private val shouldBuildNextBlock: (ConsensusRoundIdentifier) -> Boolean,
  private val blockBuilderIdentity: Validator,
) : BeaconBlockImporter {
  override fun importBlock(beaconBlock: BeaconBlock): SafeFuture<ForkChoiceUpdatedResult> {
    val beaconBlockHeader = beaconBlock.beaconBlockHeader
    val finalizationState = finalizationStateProvider(beaconBlockHeader)
    val nextBlocksRoundIdentifier = ConsensusRoundIdentifier(beaconBlockHeader.number.toLong() + 1, 0)
    return if (shouldBuildNextBlock(nextBlocksRoundIdentifier)) {
      executionLayerManager.setHeadAndStartBlockBuilding(
        headHash = beaconBlock.beaconBlockBody.executionPayload.blockHash,
        safeHash = finalizationState.safeBlockHash,
        finalizedHash = finalizationState.finalizedBlockHash,
        nextBlockTimestamp = nextBlockTimestampProvider(nextBlocksRoundIdentifier),
        feeRecipient = blockBuilderIdentity.address,
      )
    } else {
      executionLayerManager
        .setHead(
          headHash = beaconBlock.beaconBlockBody.executionPayload.blockHash,
          safeHash = finalizationState.safeBlockHash,
          finalizedHash = finalizationState.finalizedBlockHash,
        )
    }
  }
}
