/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.subscription

import java.util.concurrent.CopyOnWriteArrayList
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import tech.pegasys.teku.infrastructure.async.SafeFuture

class InOrderFanoutSubscriptionManagerTest {
  private lateinit var subscriptionManager: InOrderFanoutSubscriptionManager<String>

  @BeforeEach
  fun setUp() {
    subscriptionManager = InOrderFanoutSubscriptionManager()
  }

  @Test
  fun `should throw when handlerId is already assigned`() {
    subscriptionManager.addSyncSubscriber("handler1", { data -> "handler1 called with: $data" })
    assertThrows<IllegalArgumentException> {
      subscriptionManager.addSyncSubscriber("handler1", { data -> "handler1 called with: $data" })
    }
    assertThrows<IllegalArgumentException> {
      subscriptionManager.addAsyncSubscriber(
        "handler1",
        { data -> SafeFuture.completedFuture("handler1 called with: $data") },
      )
    }
  }

  @Test
  fun `should throw when handler is already subscribed even with different id`() {
    val syncHandler = { data: String -> "sync handler called with: $data" }
    val aSyncHandler = { data: String -> SafeFuture.completedFuture("async handler called with $data") }

    subscriptionManager.addSyncSubscriber("handler1", syncHandler)
    assertThrows<IllegalArgumentException> {
      subscriptionManager.addSyncSubscriber("handler2", syncHandler)
    }

    subscriptionManager.addAsyncSubscriber("async handler1", aSyncHandler)
    assertThrows<IllegalArgumentException> {
      subscriptionManager.addSyncSubscriber("async handler2", aSyncHandler)
    }
  }

  @Test
  fun `should notify in subscription order`() {
    val notifications = CopyOnWriteArrayList<String>()
    val numberOfHandlers = 40
    (1..numberOfHandlers).forEach { i ->
      if (i % 2 == 0) {
        subscriptionManager.addSyncSubscriber(
          "handler$i",
          { data -> notifications.add("handler$i called with: $data") },
        )
      } else {
        subscriptionManager.addAsyncSubscriber("handler$i", { data ->
          notifications.add("handler$i called with: $data")
          SafeFuture.runAsync({
            Thread.sleep(5) // Simulate async processing delay
            "handler2 called with: $data"
          })
        })
      }
    }

    subscriptionManager.notifySubscribers("d1")
    subscriptionManager.notifySubscribersAsync("d2")
    subscriptionManager.notifySubscribers("d3")
    subscriptionManager.notifySubscribersAsync("d4")

    val expectedNotifications1 = (1..numberOfHandlers).map { i -> "handler$i called with: d1" }
    val expectedNotifications2 = (1..numberOfHandlers).map { i -> "handler$i called with: d2" }
    val expectedNotifications3 = (1..numberOfHandlers).map { i -> "handler$i called with: d3" }
    val expectedNotifications4 = (1..numberOfHandlers).map { i -> "handler$i called with: d4" }

    assertThat(notifications).isEqualTo(
      expectedNotifications1 + expectedNotifications2 + expectedNotifications3 + expectedNotifications4,
    )
  }

  @Test
  fun `notifyAsync should notify subscribers without blocking calling thread`() {
    val notifications = CopyOnWriteArrayList<String>()
    subscriptionManager.addSyncSubscriber("handler1", { data -> notifications.add("handler1 called with: $data") })
    val futures = CopyOnWriteArrayList<SafeFuture<String>>()
    subscriptionManager.addAsyncSubscriber("handler2", { data ->
      val futureResult = SafeFuture<String>()
      futureResult.thenPeek { notifications.add("async handler2 called with: $data") }
      futures.add(futureResult)
      futureResult
    })

    subscriptionManager.notifySubscribersAsync("d1")
    subscriptionManager.notifySubscribersAsync("d2")
    subscriptionManager.notifySubscribersAsync("d3")
    // only handler1 is synchronous, so it should be called immediately
    // handler2 is async, so we need to wait for it
    assertThat(notifications).isEqualTo(
      listOf(
        "handler1 called with: d1",
        "handler1 called with: d2",
        "handler1 called with: d3",
      ),
    )
    futures.forEach { it.complete("handler2 completed") }

    assertThat(notifications).isEqualTo(
      listOf(
        "handler1 called with: d1",
        "handler1 called with: d2",
        "handler1 called with: d3",
        "async handler2 called with: d1",
        "async handler2 called with: d2",
        "async handler2 called with: d3",
      ),
    )
  }

  @Test
  fun `should be resilient to exceptions in subscribers`() {
    val notifications = CopyOnWriteArrayList<String>()
    subscriptionManager.addSyncSubscriber("handler1", { data ->
      notifications.add("handler1 called with: $data")
      throw RuntimeException("error in handler1")
    })
    subscriptionManager.addAsyncSubscriber("handler2", { data ->
      notifications.add("handler2 called with: $data")
      throw RuntimeException("error in handler2")
    })
    subscriptionManager.addAsyncSubscriber("handler3", { data ->
      notifications.add("handler3 called with: $data")
      SafeFuture.runAsync({
        Thread.sleep(50) // Simulate async processing delay
        throw RuntimeException("error in handler3")
      })
    })
    subscriptionManager.addSyncSubscriber("handler4", { data ->
      notifications.add("handler4 called with: $data")
    })
    subscriptionManager.addAsyncSubscriber("handler5", { data ->
      notifications.add("handler5 called with: $data")
      SafeFuture.completedFuture<String>("handler5 completed with: $data")
    })

    subscriptionManager.notifySubscribers("d1")
    subscriptionManager.notifySubscribersAsync("d2")

    assertThat(notifications).isEqualTo(
      listOf(
        "handler1 called with: d1",
        "handler2 called with: d1",
        "handler3 called with: d1",
        "handler4 called with: d1",
        "handler5 called with: d1",
        "handler1 called with: d2",
        "handler2 called with: d2",
        "handler3 called with: d2",
        "handler4 called with: d2",
        "handler5 called with: d2",
      ),
    )
  }

  @Test
  fun `should not notify removed subscribers`() {
    val notifications = CopyOnWriteArrayList<String>()

    subscriptionManager.addSyncSubscriber("handler1", { data ->
      notifications.add("handler1 called with: $data")
    })
    subscriptionManager.notifySubscribers("d1")
    assertThat(notifications).isEqualTo(listOf("handler1 called with: d1"))

    subscriptionManager.removeSubscriber("handler1")

    subscriptionManager.notifySubscribers("d2")
    assertThat(notifications).isEqualTo(listOf("handler1 called with: d1"))
  }

  @Test
  fun `should handle empty subscription list gracefully`() {
    subscriptionManager.notifySubscribers("d1")
    subscriptionManager.notifySubscribersAsync("d2")
  }
}
