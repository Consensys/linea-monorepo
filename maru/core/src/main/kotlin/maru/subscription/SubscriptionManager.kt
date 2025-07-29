/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.subscription

import tech.pegasys.teku.infrastructure.async.SafeFuture

interface SubscriptionManager<T> {
  fun addSyncSubscriber(
    subscriberId: String,
    syncHandler: (T) -> Any?,
  )

  fun addAsyncSubscriber(
    subscriberId: String,
    asyncHandler: (T) -> SafeFuture<*>,
  )

  /**
   * It will notify all subscribers in a blocking manner, by order they were added.
   * If there sync and async subscribers, it will wait for async subscriber to complete
   * before notifying the next subscriber.
   */
  fun notifySubscribers(data: T)

  /**
   * It will notify all subscribers asynchronously, without blocking the current thread.
   * Subscriber notification order my not follow the order they were added,
   */
  fun notifySubscribersAsync(data: T): SafeFuture<Unit>

  fun removeSubscriber(subscriberId: String)
}
