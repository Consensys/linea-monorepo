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

import java.time.Clock
import java.util.concurrent.TimeUnit
import kotlin.time.Duration
import maru.consensus.ForksSchedule
import maru.consensus.NextBlockTimestampProvider
import maru.core.Protocol
import maru.executionlayer.manager.BlockMetadata
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.consensus.common.bft.ConsensusRoundIdentifier
import org.hyperledger.besu.consensus.common.bft.events.BlockTimerExpiry
import org.hyperledger.besu.consensus.common.bft.statemachine.BftEventHandler
import tech.pegasys.teku.infrastructure.async.SafeFuture

class DummyConsensus(
  private val forksSchedule: ForksSchedule,
  private val eventHandler: BftEventHandler,
  private val blockMetadataProvider: () -> BlockMetadata,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val clock: Clock,
  private val config: Config,
) : Protocol {
  data class Config(
    val communicationMargin: Duration,
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  @Volatile
  private var currentTask: SafeFuture<Unit>? = null

  @Synchronized
  override fun start() {
    if (currentTask == null) {
      SafeFuture.runAsync {
        handleTick()
      }
    } else {
      throw IllegalStateException("Timer has already been started!")
    }
  }

  @Synchronized
  private fun handleTick() {
    val lastBlockMetadata = blockMetadataProvider()
    val nextTargetBlockTimestampSeconds =
      nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(lastBlockMetadata.unixTimestampSeconds)
    val nextBlockFork =
      forksSchedule.getForkByTimestamp(nextTargetBlockTimestampSeconds)
    val nextBlockNumber = lastBlockMetadata.blockNumber + 1u

    log.debug("currentTimestamp={} nextBlockNumber={}", clock.millis(), nextBlockNumber)

    if (currentTask != null) {
      if (!currentTask!!.isDone) {
        log.warn("Current task isn't done. Scheduling the next one, but results may be unexpected!")
      }
      stop()
    }

    when (nextBlockFork.configuration) {
      is DummyConsensusConfig -> {
        scheduleNextTask(
          nextBlockNumber = nextBlockNumber,
          nextTargetTimestampSeconds = nextTargetBlockTimestampSeconds,
        )
      }

      else -> {
        log.warn("Next fork isn't a Dummy Consensus one. Stopping {}", this.javaClass.name)
      }
    }
  }

  private fun scheduleNextTask(
    nextBlockNumber: ULong,
    nextTargetTimestampSeconds: Long,
  ) {
    val currentTime = clock.millis()

    val delayMillis: Long =
      nextTargetTimestampSeconds * 1000 - currentTime - config.communicationMargin.inWholeMilliseconds
    log.debug("Next target timestamp: {}, delay until next task: {}", currentTime + delayMillis, delayMillis)

    val executor =
      SafeFuture.delayedExecutor(delayMillis, TimeUnit.MILLISECONDS)

    currentTask =
      SafeFuture
        .of(
          SafeFuture
            .runAsync(
              {
                val consensusRoundIdentifier =
                  ConsensusRoundIdentifier(nextBlockNumber.toLong(), nextBlockNumber.toInt())
                log.debug("Triggering an event for round {}", consensusRoundIdentifier)
                eventHandler.handleBlockTimerExpiry(BlockTimerExpiry(consensusRoundIdentifier))
              },
              executor,
            ),
        ).handleException {
          log.error(it.message, it)
        }.thenApply {
          handleTick()
        }
  }

  @Synchronized
  override fun stop() {
    if (currentTask != null) {
      currentTask!!.cancel(false)
      currentTask = null
    } else {
      throw IllegalStateException("EventProducer hasn't been started to stop it!")
    }
  }
}
