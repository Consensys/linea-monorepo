/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
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

  override fun broadcastMessage(message: Message<*, GossipMessageType>): SafeFuture<Unit> {
    log.debug("Doing nothing for message={}", message)
    return SafeFuture.completedFuture(Unit)
  }

  override fun subscribeToBlocks(subscriber: SealedBeaconBlockHandler<ValidationResult>): Int {
    log.debug("Subscription called for subscriber={}", subscriber)
    return 0
  }

  override fun unsubscribeFromBlocks(subscriptionId: Int) {
    log.debug("Unsubscription called for subscriptionId={}", subscriptionId)
  }

  override val port: UInt
    get() = 0u
}
