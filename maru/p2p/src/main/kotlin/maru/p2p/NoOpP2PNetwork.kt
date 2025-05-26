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
package maru.p2p

import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture

object NoOpP2PNetwork : P2PNetwork {
  private val log = LogManager.getLogger(this.javaClass)

  override fun start(): SafeFuture<Unit> =
    SafeFuture
      .fromRunnable {
        log.debug("NoopP2PNetwork started")
      }.thenApply { }

  override fun stop(): SafeFuture<Unit> =
    SafeFuture
      .fromRunnable {
        log.debug("NoopP2PNetwork stopped")
      }.thenApply { }

  override fun broadcastMessage(message: Message<*>) {
    log.debug("Doing nothing for message={}", message)
  }

  override fun subscribeToBlocks(subscriber: SealedBeaconBlockHandler): Int {
    log.debug("Subscription called for subscriber={}", subscriber)
    return 0
  }

  override fun unsubscribe(subscriptionId: Int) {
    log.debug("Unsubscription called for subscriptionId={}", subscriptionId)
  }
}
