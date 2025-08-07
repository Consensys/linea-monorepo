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
import maru.consensus.state.FinalizationProvider
import maru.core.GENESIS_EXECUTION_PAYLOAD
import maru.database.BeaconChain
import maru.executionlayer.manager.ExecutionLayerManager
import maru.executionlayer.manager.ExecutionPayloadStatus
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
}

/**
 * Polls the EL for its latest block
 * If it's behind the executionPayload block number in the `beaconChain` by more than `leeway` it sends the
 * status update callback `onStatusChange` and it tries to sync the EL by `executionLayerManager` with the latest known
 * EL hash from the `beaconChain`
 */
class ELSyncService(
  private val beaconChain: BeaconChain,
  private val executionLayerManager: ExecutionLayerManager,
  private val onStatusChange: (ELSyncStatus) -> Unit,
  private val config: Config,
  private val finalizationProvider: FinalizationProvider,
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
  private var currentELSyncStatus: ELSyncStatus? = null
  private var currentElSyncTarget: ElBlockInfo? = null

  private fun pollTask() {
    val latestBeaconBlockHeader = beaconChain.getLatestBeaconState().latestBeaconBlockHeader
    val latestSealedBeaconBlock = beaconChain.getSealedBeaconBlock(beaconBlockNumber = latestBeaconBlockHeader.number)!!
    val latestBeaconBlockBody = latestSealedBeaconBlock.beaconBlock.beaconBlockBody
    val newElSyncTarget =
      ElBlockInfo(
        blockNumber = latestBeaconBlockBody.executionPayload.blockNumber,
        blockHash = latestBeaconBlockBody.executionPayload.blockHash,
      )
    if (newElSyncTarget != currentElSyncTarget) {
      log.info("New elSyncTarget={}", newElSyncTarget)
      currentElSyncTarget = newElSyncTarget
    } else {
      log.trace("Current elSyncTarget={}", currentElSyncTarget)
    }

    if (currentElSyncTarget!!.blockNumber == GENESIS_EXECUTION_PAYLOAD.blockNumber) {
      val newELSyncStatus = ELSyncStatus.SYNCED
      if (currentELSyncStatus != newELSyncStatus) {
        currentELSyncStatus = newELSyncStatus
        onStatusChange(newELSyncStatus)
      }
      return
    }

    val finalizationState = finalizationProvider(latestBeaconBlockBody)
    val fcuResponse =
      executionLayerManager
        .newPayload(latestBeaconBlockBody.executionPayload)
        .thenCompose {
          executionLayerManager
            .setHead(
              headHash = currentElSyncTarget!!.blockHash,
              safeHash = finalizationState.safeBlockHash,
              finalizedHash = finalizationState.finalizedBlockHash,
            )
        }.get()

    val newELSyncStatus =
      when (fcuResponse.payloadStatus.status) {
        ExecutionPayloadStatus.SYNCING -> {
          log.debug("EL client went out of sync")
          ELSyncStatus.SYNCING
        }

        ExecutionPayloadStatus.VALID -> {
          log.debug(
            "EL client is synced elBlockHash={} elBlockNumber={}",
            newElSyncTarget.blockHash.encodeHex(),
            latestBeaconBlockBody.executionPayload.blockNumber,
          )
          ELSyncStatus.SYNCED
        }

        else -> throw IllegalStateException("Unexpected payload status: ${fcuResponse.payloadStatus.status}")
      }

    if (currentELSyncStatus != newELSyncStatus) {
      currentELSyncStatus = newELSyncStatus
      onStatusChange(newELSyncStatus)
    }
  }

  override fun start() {
    synchronized(this) {
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
            log.warn("ELSyncService poll task exception", e)
          }
        },
        0,
        config.pollingInterval.inWholeMilliseconds,
      )
    }
  }

  override fun stop() {
    synchronized(this) {
      if (poller == null) {
        return
      }
      poller?.cancel()
      poller = null
      currentELSyncStatus = null
    }
  }
}
