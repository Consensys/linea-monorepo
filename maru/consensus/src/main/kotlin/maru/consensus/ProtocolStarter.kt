/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus

import java.time.Clock
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds
import linea.timer.Timer
import linea.timer.TimerFactory
import maru.core.Protocol
import maru.subscription.SubscriptionNotifier
import maru.syncing.CLSyncStatus
import maru.syncing.SyncStatusProvider
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class ProtocolStarter(
  private val forksSchedule: ForksSchedule,
  private val protocolFactory: ProtocolFactory,
  private val nextBlockTimestampProvider: NextBlockTimestampProvider,
  private val forkTransitionCheckInterval: Duration,
  private val clock: Clock = Clock.systemUTC(),
  private val timerFactory: TimerFactory,
  private val forkTransitionNotifier: SubscriptionNotifier<ForkSpec>,
) : Protocol {
  companion object {
    fun create(
      forksSchedule: ForksSchedule,
      protocolFactory: ProtocolFactory,
      nextBlockTimestampProvider: NextBlockTimestampProvider,
      syncStatusProvider: SyncStatusProvider,
      forkTransitionCheckInterval: Duration = 1.seconds,
      clock: Clock = Clock.systemUTC(),
      timerFactory: TimerFactory,
      forkTransitionNotifier: SubscriptionNotifier<ForkSpec>,
    ): ProtocolStarter {
      val protocolStarter =
        ProtocolStarter(
          forksSchedule = forksSchedule,
          protocolFactory = protocolFactory,
          nextBlockTimestampProvider = nextBlockTimestampProvider,
          forkTransitionCheckInterval = forkTransitionCheckInterval,
          clock = clock,
          timerFactory = timerFactory,
          forkTransitionNotifier = forkTransitionNotifier,
        )
      syncStatusProvider.onClSyncStatusUpdate {
        if (it == CLSyncStatus.SYNCING) {
          protocolStarter.pause()
        }
      }
      syncStatusProvider.onBeaconSyncComplete {
        try {
          protocolStarter.start()
        } catch (th: Throwable) {
          throw th
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

  private val log: Logger = LogManager.getLogger(this.javaClass)

  internal val currentProtocolWithForkReference: AtomicReference<ProtocolWithFork> = AtomicReference()

  private var poller: Timer =
    timerFactory.createTimer(
      name = "ProtocolStarterPoller",
      initialDelay = forkTransitionCheckInterval,
      period = forkTransitionCheckInterval,
      timerSchedule = linea.timer.TimerSchedule.FIXED_RATE,
      errorHandler = {},
      task = Runnable { pollTask() },
    )

  private fun pollTask() {
    try {
      checkAndHandleForkTransition()
    } catch (th: Throwable) {
      log.error("Error during fork transition check", th)
    }
  }

  private fun checkAndHandleForkTransition() {
    val currentTimestamp = clock.instant().epochSecond.toULong()
    val nextBlockTimestamp = nextBlockTimestampProvider.nextTargetBlockUnixTimestamp(currentTimestamp)
    val nextForkSpec = forksSchedule.getForkByTimestamp(nextBlockTimestamp)

    val currentProtocolWithFork = currentProtocolWithForkReference.get()

    if (currentProtocolWithFork?.fork != nextForkSpec) {
      log.debug(
        "switching from forkSpec={} to newForkSpec={}, nextBlockTimeStamp={}",
        currentProtocolWithFork?.fork,
        nextForkSpec,
        nextBlockTimestamp,
      )

      performForkTransition(currentProtocolWithFork, nextForkSpec)
    } else {
      log.trace("currentTimestamp={}, but fork switch isn't required", currentTimestamp)
    }
  }

  @Synchronized
  private fun performForkTransition(
    currentProtocolWithFork: ProtocolWithFork?,
    nextForkSpec: ForkSpec,
  ) {
    val newProtocol: Protocol = protocolFactory.create(nextForkSpec)

    val newProtocolWithFork =
      ProtocolWithFork(
        newProtocol,
        nextForkSpec,
      )
    log.debug("switching protocol: fromProtocol={} toProtocol={}", currentProtocolWithFork, newProtocolWithFork)
    currentProtocolWithForkReference.set(newProtocolWithFork)
    currentProtocolWithFork?.protocol?.close()

    newProtocol.start()
    log.debug("started new protocol {}", newProtocol)
    forkTransitionNotifier.notifySubscribers(nextForkSpec)
  }

  override fun start() {
    synchronized(this) {
      checkAndHandleForkTransition()
      poller.start()
      log.debug("Starting fork transition polling with interval {}", forkTransitionCheckInterval)
    }
  }

  override fun pause() {
    synchronized(this) {
      poller.stop()
      currentProtocolWithForkReference.get()?.protocol?.pause()
      log.debug("Stopped fork transition polling")
    }
  }

  override fun close() {
    synchronized(this) {
      pause()
      currentProtocolWithForkReference.get()?.protocol?.close()
    }
  }
}
