/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.qbft

import kotlin.time.Duration
import maru.consensus.qbft.adapters.toBeaconBlockHeader
import maru.consensus.state.FinalizationProvider
import maru.core.Validator
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
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
  private val finalizationStateProvider: FinalizationProvider,
  private val blockBuilderIdentity: Validator,
  private val beaconChain: BeaconChain,
  private val config: Config,
) : QbftBlockCreator {
  private val log: Logger = LogManager.getLogger(this.javaClass)

  data class Config(
    val minBlockBuildTime: Duration,
  )

  override fun createBlock(
    headerTimeStampSeconds: Long,
    parentHeader: QbftBlockHeader,
  ): QbftBlock {
    val beaconBlockHeader = parentHeader.toBeaconBlockHeader()
    val parentBeaconBlockBody =
      beaconChain
        .getSealedBeaconBlock(beaconBlockHeader.hash)
        ?.beaconBlock
        ?.beaconBlockBody
        ?: throw IllegalStateException("Parent block not found in the database")
    val finalizedState = finalizationStateProvider(parentBeaconBlockBody)
    val blockBuildingTriggerResult =
      manager
        .setHeadAndStartBlockBuilding(
          headHash = parentBeaconBlockBody.executionPayload.blockHash,
          safeHash = finalizedState.safeBlockHash,
          finalizedHash = finalizedState.finalizedBlockHash,
          nextBlockTimestamp = headerTimeStampSeconds,
          feeRecipient = blockBuilderIdentity.address,
        ).get()
    log.debug(
      "Building new block, FCU result={}",
      blockBuildingTriggerResult,
    )
    val sleepTime = config.minBlockBuildTime.inWholeMilliseconds
    log.debug("Block building has started, sleeping for {} milliseconds", sleepTime)
    Thread.sleep(sleepTime)
    log.debug("Block building has finished, time to collect block building results")
    return delegate.createBlock(headerTimeStampSeconds, parentHeader)
  }

  override fun createSealedBlock(
    block: QbftBlock,
    roundNumber: Int,
    commitSeals: Collection<SECPSignature>,
  ): QbftBlock = DelayedQbftBlockCreator.createSealedBlock(block, roundNumber, commitSeals)
}
