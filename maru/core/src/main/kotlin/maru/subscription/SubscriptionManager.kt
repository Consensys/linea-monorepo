/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.subscription

import java.util.concurrent.CompletableFuture
import java.util.concurrent.CopyOnWriteArrayList
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface SubscriptionManager<T> {
  fun addSyncSubscriber(
    subscriberId: String,
    syncHandler: (T) -> Any?,
  )

  fun addAsyncSubscriber(
    subscriberId: String,
    asyncHandler: (T) -> CompletableFuture<*>,
  )

  /**
   * It will notify all subscribers in a blocking manner, by order they were added.
   * If there sync and async subscribers, it will wait for async subscriber to complete
   * before notifying the next subscriber.
   */
  fun notifySubscribers(data: T)

  /**
   * It will notify all subscribers asynchronously, without blocking the current thread.
   * Subscribers notification order my not follow the order they were added.
   *
   * If the called does not wait on Future resolution,
   * the Order of notifications is not guaranteed, and may depend on the concrete implementation.
   *
   * Returned when all subscribers have been notified and resolved their futures.
   * If any subscriber fails, the future will be completed exceptionally.
   */
  fun notifySubscribersAsync(data: T): SafeFuture<Unit>

  fun removeSubscriber(subscriberId: String)
}

/**
 * This is a simple implementation of SubscriptionManager that notifies all subscribers
 * in a fanout manner. Subscribers are notified in the order they are added.
 */
class InOrderFanoutSubscriptionManager<T>(
  val log: Logger = LogManager.getLogger(InOrderFanoutSubscriptionManager::class.java),
) : SubscriptionManager<T> {
  private data class SubscriberData<T>(
    val id: String,
    val syncHandler: ((T) -> Any?)? = null,
    val asyncHandler: ((T) -> CompletableFuture<*>)? = null,
  ) {
    init {
      require(syncHandler != null || asyncHandler != null) {
        "Either syncHandler or asyncHandler must be provided"
      }
      require(syncHandler == null || asyncHandler == null) {
        "Only one of syncHandler or asyncHandler can be provided"
      }
    }
  }

  private val subscribers: MutableList<SubscriberData<T>> = CopyOnWriteArrayList()

  private fun ensureUniqueSubscriber(
    subscriberId: String,
    handler: Any?,
  ) {
    subscribers
      .find { it.id == subscriberId || it.syncHandler == handler || it.asyncHandler == handler }
      ?.let { subscriber ->
        throw IllegalArgumentException("handler already subscribed with subscriberId=${subscriber.id}")
      }
  }

  override fun addSyncSubscriber(
    subscriberId: String,
    syncHandler: (T) -> Any?,
  ) {
    ensureUniqueSubscriber(subscriberId, syncHandler)
    subscribers.add(SubscriberData(subscriberId, syncHandler))
  }

  override fun addAsyncSubscriber(
    subscriberId: String,
    asyncHandler: (T) -> CompletableFuture<*>,
  ) {
    ensureUniqueSubscriber(subscriberId, asyncHandler)
    subscribers.add(SubscriberData(subscriberId, asyncHandler = asyncHandler))
  }

  override fun notifySubscribers(data: T) {
    logIfEmptySubscribers(data)
    subscribers.forEach { subscriber ->
      try {
        subscriber.syncHandler?.invoke(data)
          ?: subscriber.asyncHandler?.invoke(data)?.join()
      } catch (th: Throwable) {
        logHandlingError(
          subscriberId = subscriber.id,
          data = data,
          th = th,
        )
      }
    }
  }

  override fun notifySubscribersAsync(data: T): SafeFuture<Unit> {
    logIfEmptySubscribers(data)
    val futures =
      subscribers.map { subscriber ->
        try {
          subscriber.syncHandler
            ?.let { SafeFuture.completedFuture(it.invoke(data)) }
            ?: subscriber.asyncHandler?.invoke(data)!!
        } catch (th: Throwable) {
          logHandlingError(
            subscriberId = subscriber.id,
            data = data,
            th = th,
          )
          SafeFuture.failedFuture<Unit>(th)
        }
      }
    return SafeFuture.of(CompletableFuture.allOf(*futures.toTypedArray())).thenApply { }
  }

  override fun removeSubscriber(subscriberId: String) {
    val removed = subscribers.removeIf { it.id == subscriberId }
    if (!removed) {
      log.warn("No subscriber found with subscriberId={}", subscriberId)
    }
  }

  private fun logIfEmptySubscribers(data: T) {
    if (subscribers.isEmpty()) {
      log.warn("empty subscribers list, following data wont be delivered: data={}", data)
    }
  }

  private fun logHandlingError(
    subscriberId: String,
    data: T,
    th: Throwable,
  ) {
    log.error(
      "errorMessage={} from subscriber={} handling data={}",
      th.message ?: "Unknown error",
      subscriberId,
      data,
      th,
    )
  }
}
