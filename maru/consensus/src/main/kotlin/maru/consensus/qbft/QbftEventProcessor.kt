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

import java.util.Optional
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit
import org.apache.logging.log4j.LogManager
import org.hyperledger.besu.consensus.common.bft.BftEventQueue
import org.hyperledger.besu.consensus.common.bft.events.BftEvent

class QbftEventProcessor(
  private val incomingQueue: BftEventQueue,
  private val eventMultiplexer: QbftEventMultiplexer,
) : Runnable {
  private val log: org.apache.logging.log4j.Logger = LogManager.getLogger(this::class.java)
  private val shutdownLatch = CountDownLatch(1)

  @Volatile private var shutdown = false

  /**
   * Indicate to the processor that it can be started
   */
  @Synchronized
  fun start() {
    shutdown = false
  }

  /**
   * Indicate to the processor that it should gracefully stop at its next opportunity
   */
  @Synchronized
  fun stop() {
    shutdown = true
  }

  /**
   * Await stop.
   *
   * @throws InterruptedException the interrupted exception
   */
  @Throws(InterruptedException::class)
  fun awaitStop() {
    shutdownLatch.await()
  }

  override fun run() {
    try {
      // Start the event queue. Until it is started it won't accept new events from peers
      incomingQueue.start()

      while (!shutdown) {
        nextEvent().ifPresent { eventMultiplexer.handleEvent(it) }
      }

      incomingQueue.stop()
    } catch (t: Throwable) {
      log.error("BFT Mining thread has suffered a fatal error, mining has been halted", t)
    }

    // Clean up the executor service the round timer has been utilising
    log.info("Shutting down BFT event processor")
    shutdownLatch.countDown()
  }

  private fun nextEvent(): Optional<BftEvent> =
    try {
      Optional.ofNullable(incomingQueue.poll(500, TimeUnit.MILLISECONDS))
    } catch (_: InterruptedException) {
      // If the queue was interrupted propagate it and spin to check our shutdown status
      Thread.currentThread().interrupt()
      Optional.empty()
    }
}
