/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import java.util.concurrent.atomic.AtomicReference
import kotlin.math.max
import maru.core.BeaconBlock
import maru.core.Protocol
import maru.syncing.ELSyncStatus
import maru.syncing.SyncStatusProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ProtocolStarterBlockHandler(
  private val protocolStarter: ProtocolStarter,
) : NewBlockHandler<Unit> {
  override fun handleNewBlock(beaconBlock: BeaconBlock): SafeFuture<Unit> {
    val elBlockMetadata =
      ElBlockMetadata(
        beaconBlock.beaconBlockBody.executionPayload.blockNumber,
        beaconBlock.beaconBlockHeader.hash,
        beaconBlock.beaconBlockHeader.timestamp.toLong(),
      )
    protocolStarter.handleNewElBlock(elBlockMetadata)
    return SafeFuture.completedFuture(Unit)
  }
}

class ProtocolStarter(
  private val forksSchedule: ForksSchedule,
  private val protocolFactory: ProtocolFactory,
  private val elMetadataProvider: ElMetadataProvider, // TODO: we should probably replace it with BeaconChain
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
) : Protocol {
  companion object {
    fun create(
      forksSchedule: ForksSchedule,
      protocolFactory: ProtocolFactory,
      elMetadataProvider: ElMetadataProvider,
      nextBlockTimestampProvider: NextBlockTimestampProvider,
      syncStatusProvider: SyncStatusProvider,
    ): ProtocolStarter {
      val protocolStarter =
        ProtocolStarter(
          forksSchedule = forksSchedule,
          protocolFactory = protocolFactory,
          elMetadataProvider = elMetadataProvider,
          nextBlockTimestampProvider = nextBlockTimestampProvider,
        )
      syncStatusProvider.onElSyncStatusUpdate {
        when (it) {
          ELSyncStatus.SYNCING -> protocolStarter.stop()
          ELSyncStatus.SYNCED -> {
            try {
              protocolStarter.start()
            } catch (th: Throwable) {
              throw th
            }
          }
        }
      }
      return protocolStarter
    }
  }

  data class ProtocolWithFork(
    val protocol: Protocol,
    val fork: ForkSpec,
  ) {
    override fun toString(): String = "protocol=${protocol.javaClass.simpleName}, fork=$fork"
  }

  private val log: Logger = LogManager.getLogger(this::class.java)

  internal val currentProtocolWithForkReference: AtomicReference<ProtocolWithFork> = AtomicReference()

  @Synchronized
  fun handleNewElBlock(elBlock: ElBlockMetadata) {
    log.debug("New blockNumber={} received", { elBlock.blockNumber })

    val nextBlockTimestamp = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(elBlock.unixTimestampSeconds)
    val nextForkSpec = forksSchedule.getForkByTimestamp(nextBlockTimestamp)

    val currentProtocolWithFork = currentProtocolWithForkReference.get()
    if (currentProtocolWithFork?.fork != nextForkSpec) {
      log.debug("switching from forkSpec={} to newForkSpec={}", currentProtocolWithFork?.fork, nextForkSpec)
      val newProtocol: Protocol = protocolFactory.create(nextForkSpec)

      val newProtocolWithFork =
        ProtocolWithFork(
          newProtocol,
          nextForkSpec,
        )
      log.debug("switched protocol: fromProtocol={} toProtocol={}", currentProtocolWithFork, newProtocolWithFork)
      currentProtocolWithForkReference.set(
        newProtocolWithFork,
      )
      currentProtocolWithFork?.protocol?.stop()

      // Wait until timestamp is reached before starting new protocol
      val timeTillFork = max((nextBlockTimestamp * 1000L) - System.currentTimeMillis(), 0)
      log.debug("Waiting for {} ms until fork switch", timeTillFork)
      Thread.sleep(timeTillFork)

      newProtocol.start()
      log.debug("stated new protocol {}", newProtocol)
    } else {
      log.trace("block {} was produced, but the fork switch isn't required", { elBlock.blockNumber })
    }
  }

  override fun start() {
    val latestBlock = elMetadataProvider.getLatestBlockMetadata()
    handleNewElBlock(latestBlock)
  }

  override fun stop() {
    currentProtocolWithForkReference.get()?.protocol?.stop()
    currentProtocolWithForkReference.set(null)
  }
}
