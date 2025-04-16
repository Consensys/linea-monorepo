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
package maru.consensus.dummy

import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.ExecutionPayload
import maru.core.HashUtil
import maru.core.HeaderHashFunction
import maru.core.Validator
import maru.executionlayer.manager.ExecutionLayerManager
import maru.serialization.rlp.KeccakHasher
import maru.serialization.rlp.RLPSerializers
import org.apache.logging.log4j.LogManager

/**
 * Responsible for block creation with Engine API. createEmptyWithdrawalsBlock is the only available signature, the
 * rest of the methods delegate to it
 * Even though it implements BlockCreator, this interface doesn't really fit Engine API flow due to its asynchronicity
 */
class DummyEngineApiBlockCreator(
  private val manager: ExecutionLayerManager,
  private val state: DummyConsensusState,
  nextBlockTimestamp: Long, // Block creation starts right after, so it's required
  private val feeRecipientProvider: FeeRecipientProvider,
  private val hasher: HeaderHashFunction =
    HashUtil.headerHash(
      RLPSerializers.BeaconBlockHeaderSerializer,
      KeccakHasher,
    ),
) {
  init {
    manager
      .setHeadAndStartBlockBuilding(
        headHash = state.latestBlockHash,
        safeHash = state.finalizationState.safeBlockHash,
        finalizedHash = state.finalizationState.finalizedBlockHash,
        nextBlockTimestamp = nextBlockTimestamp,
        feeRecipient = feeRecipientProvider.getFeeRecipient(nextBlockTimestamp),
      ).get()
  }

  private val log = LogManager.getLogger(DummyEngineApiBlockCreator::class.java)

  // Starts creation of the next block with timestamp, returns current block being built
  fun createBlock(timestamp: Long): BeaconBlock? {
    val blockBuildingResult =
      try {
        manager
          .finishBlockBuilding()
          .thenApply {
            state.updateLatestStatus(it.blockHash)
            log.info("Updating latest state")
            it
          }.get()
      } catch (e: Exception) {
        log.warn("Error during block building finish! Starting new attempt!", e)
        null
      }

    val newHeadHash = state.latestBlockHash
    // Mind the atomicity of finalization updates
    val finalizationState = state.finalizationState
    log.debug(" {}", newHeadHash)
    manager
      .setHeadAndStartBlockBuilding(
        headHash = newHeadHash,
        safeHash = finalizationState.safeBlockHash,
        finalizedHash = finalizationState.finalizedBlockHash,
        nextBlockTimestamp = timestamp,
        feeRecipient = feeRecipientProvider.getFeeRecipient(timestamp),
      ).whenException {
        log.error("Error while initiating block building!", it)
      }.get()
    return blockBuildingResult?.let {
      mapExecutionPayloadToBlock(blockBuildingResult)
    }
  }

  private fun mapExecutionPayloadToBlock(executionPayload: ExecutionPayload): BeaconBlock {
    val beaconBlockBody = BeaconBlockBody(prevCommitSeals = emptyList(), executionPayload = executionPayload)

    val beaconBlockHeader =
      BeaconBlockHeader(
        number = 0u,
        round = 0u,
        timestamp = executionPayload.timestamp,
        proposer = Validator(executionPayload.feeRecipient),
        parentRoot = ByteArray(32),
        stateRoot = ByteArray(32),
        bodyRoot = ByteArray(32),
        headerHashFunction = hasher,
      )
    return BeaconBlock(beaconBlockHeader, beaconBlockBody)
  }
}
