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

import java.util.concurrent.atomic.AtomicInteger
import java.util.function.Supplier
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture

class SubscriptionManager<E> {
  private val log = LogManager.getLogger(this.javaClass)
  private val nextSubscriptionId = AtomicInteger()
  private val subscriptions: MutableMap<Int, (E) -> SafeFuture<ValidationResult>> = mutableMapOf()

  fun hasSubscriptions(): Boolean = subscriptions.isNotEmpty()

  @Synchronized
  fun subscribeToBlocks(subscriber: (E) -> SafeFuture<ValidationResult>): Int {
    val subscriptionId = nextSubscriptionId.get()
    subscriptions[subscriptionId] = subscriber
    nextSubscriptionId.incrementAndGet()
    return subscriptionId
  }

  @Synchronized
  fun unsubscribe(subscriptionId: Int) {
    subscriptions.remove(subscriptionId)
  }

  fun handleEvent(event: E): SafeFuture<ValidationResult> {
    val handlerFutures =
      subscriptions.map { (subscriptionId, handler) ->
        try {
          handler(event)
        } catch (th: Throwable) {
          log.debug(
            Supplier<String> { "Error from subscription=$subscriptionId while handling event=$event!" },
            th,
          )
          SafeFuture.completedFuture<ValidationResult>(
            ValidationResult.Companion.Invalid(
              "Exception during event handling",
              th,
            ),
          )
        }
      }
    return if (subscriptions.isNotEmpty()) {
      SafeFuture.collectAll(handlerFutures.stream()).thenApply {
        it.reduce { acc: ValidationResult, next: ValidationResult ->
          when {
            acc is ValidationResult.Companion.Invalid -> acc
            next is ValidationResult.Companion.Invalid -> next
            acc is ValidationResult.Companion.Ignore -> acc
            next is ValidationResult.Companion.Ignore -> next
            else -> acc
          }
        }
      }
    } else {
      SafeFuture.completedFuture(
        ValidationResult.Companion.Ignore(
          "No subscription to imply message validity",
        ),
      )
    }
  }
}
