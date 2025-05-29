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

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture

class SubscriptionManagerTest {
  private lateinit var manager: SubscriptionManager<String>

  @BeforeEach
  fun setUp() {
    manager = SubscriptionManager()
  }

  @Test
  fun `hasSubscriptions returns false when empty and true after subscribe`() {
    assertThat(manager.hasSubscriptions()).isFalse()

    manager.subscribeToBlocks { SafeFuture.completedFuture(ValidationResult.Companion.Valid) }
    assertThat(manager.hasSubscriptions()).isTrue()
  }

  @Test
  fun `subscribeToBlocks returns unique ids and can unsubscribe`() {
    val id1 = manager.subscribeToBlocks { SafeFuture.completedFuture(ValidationResult.Companion.Valid) }
    val id2 = manager.subscribeToBlocks { SafeFuture.completedFuture(ValidationResult.Companion.Valid) }
    assertThat(id2).isNotEqualTo(id1)

    manager.unsubscribe(id1)
    assertThat(manager.hasSubscriptions()).isTrue()
    manager.unsubscribe(id2)
    assertThat(manager.hasSubscriptions()).isFalse()
  }

  @Test
  fun `handleEvent calls all subscribers and aggregates ValidationResult`() {
    val results = mutableListOf<String>()
    manager.subscribeToBlocks {
      results.add("first")
      SafeFuture.completedFuture(ValidationResult.Companion.Valid)
    }
    manager.subscribeToBlocks {
      results.add("second")
      SafeFuture.completedFuture(ValidationResult.Companion.Ignore("meh"))
    }
    val future = manager.handleEvent("event")
    val result = future.get()
    assertThat(results).containsExactly("first", "second")
    assertThat(result).isInstanceOf(ValidationResult.Companion.Ignore::class.java)
  }

  @Test
  fun `handleEvent returns Failed if any subscriber fails`() {
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Valid)
    }
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Invalid("fail"))
    }
    val result = manager.handleEvent("event").get()
    assertThat(result).isInstanceOf(ValidationResult.Companion.Invalid::class.java)
  }

  @Test
  fun `handleEvent returns KindaFine if no subscriptions`() {
    val result = manager.handleEvent("event").get()
    assertThat(result).isInstanceOf(ValidationResult.Companion.Ignore::class.java)
  }

  @Test
  fun `handleEvent handles subscriber exceptions gracefully`() {
    manager.subscribeToBlocks {
      throw RuntimeException("boom")
    }
    val result = manager.handleEvent("event").get()
    assertThat(result).isInstanceOf(ValidationResult.Companion.Invalid::class.java)
  }

  @Test
  fun `handleEvent aggregates with Failed over KindaFine and Successful`() {
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Valid)
    }
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Ignore("meh"))
    }
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Invalid("fail"))
    }
    val result = manager.handleEvent("event").get()
    assertThat(result).isInstanceOf(ValidationResult.Companion.Invalid::class.java)
  }

  @Test
  fun `handleEvent aggregates with KindaFine over Successful`() {
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Valid)
    }
    manager.subscribeToBlocks {
      SafeFuture.completedFuture(ValidationResult.Companion.Ignore("meh"))
    }
    val result = manager.handleEvent("event").get()
    assertThat(result).isInstanceOf(ValidationResult.Companion.Ignore::class.java)
  }

  @Test
  fun `only subscribed handler is called after unsubscription`() {
    var called1 = false
    var called2 = false
    val id1 =
      manager.subscribeToBlocks {
        called1 = true
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    manager.subscribeToBlocks {
      called2 = true
      SafeFuture.completedFuture(ValidationResult.Companion.Valid)
    }
    manager.unsubscribe(id1)
    manager.handleEvent("event").get()
    assertThat(called1).isFalse()
    assertThat(called2).isTrue()
  }
}
