/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.syncing

import java.lang.Exception
import java.util.Timer
import java.util.UUID
import kotlin.concurrent.timerTask
import kotlin.time.Duration
import maru.consensus.NewBlockHandler
import maru.core.GENESIS_EXECUTION_PAYLOAD
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionPayloadStatus
import maru.executionlayer.manager.ForkChoiceUpdatedResult
import maru.extensions.encodeHex
import maru.services.LongRunningService
import org.apache.logging.log4j.LogManager

data class ElBlockInfo(
  val blockNumber: ULong,
  val blockHash: ByteArray,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as ElBlockInfo

    if (blockNumber != other.blockNumber) return false
    if (!blockHash.contentEquals(other.blockHash)) return false

    return true
  }

  override fun hashCode(): Int {
    var result = blockNumber.hashCode()
    result = 31 * result + blockHash.contentHashCode()
    return result
  }

  override fun toString(): String = "ElBlockInfo(elBlockNumber=$blockNumber, elBlockHash=${blockHash.encodeHex()})"
}

/**
 * Polls the EL for its latest block
 * If it's behind the executionPayload block number in the `beaconChain` by more than `leeway` it sends the
 * status update callback `onStatusChange` and it tries to sync the EL by `executionLayerManager` with the latest known
 * EL hash from the `beaconChain`
 */
class ELSyncService(
  private val beaconChain: BeaconChain,
  private val eLValidatorBlockImportHandler: NewBlockHandler<ForkChoiceUpdatedResult>,
  private val followerELBLockImportHandler: NewBlockHandler<Unit>,
  private val onStatusChange: (ELSyncStatus) -> Unit,
  private val config: Config,
  private val timerFactory: (String, Boolean) -> Timer = { name, isDaemon ->
    Timer(
      "$name-${UUID.randomUUID()}",
      isDaemon,
    )
  },
) : LongRunningService {
  data class Config(
    val pollingInterval: Duration,
  )

  private val log = LogManager.getLogger(this.javaClass)

  private var poller: Timer? = null
  private var currentElSyncTarget: ElBlockInfo? = null

  private fun pollTask() {
    val latestBeaconBlockHeader = beaconChain.getLatestBeaconState().beaconBlockHeader
    val latestSealedBeaconBlock = beaconChain.getSealedBeaconBlock(beaconBlockNumber = latestBeaconBlockHeader.number)!!
    val latestBeaconBlockBody = latestSealedBeaconBlock.beaconBlock.beaconBlockBody
    val newElSyncTarget =
      ElBlockInfo(
        blockNumber = latestBeaconBlockBody.executionPayload.blockNumber,
        blockHash = latestBeaconBlockBody.executionPayload.blockHash,
      )
    if (newElSyncTarget != currentElSyncTarget) {
      log.debug("New elSyncTarget={}", newElSyncTarget)
      currentElSyncTarget = newElSyncTarget
    } else {
      log.trace("Current elSyncTarget={}", currentElSyncTarget)
    }

    if (currentElSyncTarget!!.blockNumber == GENESIS_EXECUTION_PAYLOAD.blockNumber) {
      val newELSyncStatus = ELSyncStatus.SYNCED
      onStatusChange(newELSyncStatus)
      return
    }
    val fcuResponse = eLValidatorBlockImportHandler.handleNewBlock(latestSealedBeaconBlock.beaconBlock).get()
    // Call and forget, best effort
    followerELBLockImportHandler.handleNewBlock(latestSealedBeaconBlock.beaconBlock).whenException { ex ->
      log.warn(
        "Block import to followers failed: clBlockNumber={}, elBlockNumber={} errorMessage={}",
        latestBeaconBlockHeader.number,
        latestBeaconBlockBody.executionPayload.blockNumber,
        ex.message,
        ex,
      )
    }

    val newELSyncStatus =
      when (fcuResponse.payloadStatus.status) {
        ExecutionPayloadStatus.SYNCING -> {
          log.debug("EL client went out of sync")
          ELSyncStatus.SYNCING
        }

        ExecutionPayloadStatus.VALID -> {
          log.debug(
            "EL client is synced newSyncTarget={}",
            newElSyncTarget,
          )
          val latestClBlockNumber = beaconChain.getLatestBeaconState().beaconBlockHeader.number
          val latestElBlockNumber =
            beaconChain
              .getSealedBeaconBlock(beaconBlockNumber = latestClBlockNumber)!!
              .beaconBlock.beaconBlockBody.executionPayload.blockNumber
          if (newElSyncTarget.blockNumber == latestElBlockNumber) {
            ELSyncStatus.SYNCED
          } else {
            ELSyncStatus.SYNCING
          }
        }

        else -> throw IllegalStateException("Unexpected payload status: ${fcuResponse.payloadStatus.status}")
      }

    onStatusChange(newELSyncStatus)
  }

  @Synchronized
  override fun start() {
    if (poller != null) {
      return
    }
    log.debug("Starting ELSyncService with polling interval: {}", config.pollingInterval)
    poller = timerFactory("ELSyncPoller", true)
    poller!!.scheduleAtFixedRate(
      timerTask {
        try {
          pollTask()
        } catch (e: Exception) {
          log.warn("Poll task failed: errorMessage={}", e.message, e)
        }
      },
      0,
      config.pollingInterval.inWholeMilliseconds,
    )
  }

  @Synchronized
  override fun stop() {
    if (poller == null) {
      return
    }
    poller?.cancel()
    poller = null
  }
}
