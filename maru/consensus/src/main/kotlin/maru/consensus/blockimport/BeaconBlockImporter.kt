/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.blockimport

import maru.consensus.NewBlockHandler
import maru.consensus.NextBlockTimestampProvider
import maru.consensus.state.FinalizationState
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconState
import maru.core.Validator
import maru.executionlayer.client.ExecutionLayerEngineApiClient
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import maru.executionlayer.manager.JsonRpcExecutionLayerManager
import maru.p2p.ValidationResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface BeaconBlockImporter {
  fun importBlock(
    beaconState: BeaconState,
    beaconBlock: BeaconBlock,
  ): SafeFuture<*>
}

class FollowerBeaconBlockImporter(
  private val executionLayerManager: ExecutionLayerManager,
  private val finalizationStateProvider: (BeaconBlockBody) -> FinalizationState,
) : NewBlockHandler<ValidationResult> {
  companion object {
    fun create(executionLayerEngineApiClient: ExecutionLayerEngineApiClient): NewBlockHandler<ValidationResult> {
      val executionLayerManager =
        JsonRpcExecutionLayerManager(
          executionLayerEngineApiClient = executionLayerEngineApiClient,
        )
      return FollowerBeaconBlockImporter(
        executionLayerManager = executionLayerManager,
        finalizationStateProvider = {
          FinalizationState(
            it.executionPayload.blockHash,
            it.executionPayload.blockHash,
          )
        },
      )
    }
  }

  private val log = LogManager.getLogger(this.javaClass)

  override fun handleNewBlock(beaconBlock: BeaconBlock): SafeFuture<ValidationResult> {
    val executionPayload = beaconBlock.beaconBlockBody.executionPayload
    return executionLayerManager
      .newPayload(executionPayload)
      .handleException { e ->
        log.error(
          "Error importing execution payload for blockNumber=${executionPayload.blockNumber}",
          e,
        )
      }.thenCompose {
        val finalizationState = finalizationStateProvider(beaconBlock.beaconBlockBody)
        executionLayerManager
          .setHead(
            headHash = beaconBlock.beaconBlockBody.executionPayload.blockHash,
            safeHash = finalizationState.safeBlockHash,
            finalizedHash = finalizationState.finalizedBlockHash,
          ).thenApply {
            ValidationResult.fromForkChoiceUpdatedResult(it)
          }
      }
  }
}

class BlockBuildingBeaconBlockImporter(
  private val executionLayerManager: ExecutionLayerManager,
  private val finalizationStateProvider: (BeaconBlockBody) -> FinalizationState,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val shouldBuildNextBlock: (BeaconState, ConsensusRoundIdentifier) -> Boolean,
  private val blockBuilderIdentity: Validator,
) : BeaconBlockImporter {
  private val log: Logger = LogManager.getLogger(this::javaClass)

  override fun importBlock(
    beaconState: BeaconState,
    beaconBlock: BeaconBlock,
  ): SafeFuture<ForkChoiceUpdatedResult> {
    val beaconBlockHeader = beaconBlock.beaconBlockHeader
    val finalizationState = finalizationStateProvider(beaconBlock.beaconBlockBody)
    val nextBlocksRoundIdentifier = ConsensusRoundIdentifier(beaconBlockHeader.number.toLong() + 1, 0)
    return if (shouldBuildNextBlock(beaconState, nextBlocksRoundIdentifier)) {
      val nextBlockTimestamp =
        nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(
          beaconState
            .latestBeaconBlockHeader.timestamp
            .toLong(),
        )
      log.debug(
        "Importing blockHeader={} with timestamp={} and starting building of next block with timestamp={}",
        beaconBlockHeader,
        beaconBlock.beaconBlockBody.executionPayload.timestamp,
        nextBlockTimestamp,
      )
      executionLayerManager.setHeadAndStartBlockBuilding(
        headHash = beaconBlock.beaconBlockBody.executionPayload.blockHash,
        safeHash = finalizationState.safeBlockHash,
        finalizedHash = finalizationState.finalizedBlockHash,
        nextBlockTimestamp = nextBlockTimestamp,
        feeRecipient = blockBuilderIdentity.address,
      )
    } else {
      log.debug("Importing blockHeader={}", beaconBlockHeader)
      executionLayerManager.setHead(
        headHash = beaconBlock.beaconBlockBody.executionPayload.blockHash,
        safeHash = finalizationState.safeBlockHash,
        finalizedHash = finalizationState.finalizedBlockHash,
      )
    }
  }
}
