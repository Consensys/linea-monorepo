/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.p2p.topics

import java.util.Optional
import maru.p2p.MaruPreparedGossipMessage
import maru.p2p.SubscriptionManager
import maru.p2p.ValidationResult
import maru.serialization.Deserializer
import org.apache.tuweni.bytes.Bytes
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import tech.pegasys.teku.networking.p2p.gossip.PreparedGossipMessage
import io.libp2p.core.pubsub.ValidationResult as Libp2pValidationResult

class TopicHanlderWithInOrderDeliveringTest {
  private fun makeHandler(
    initialSeq: ULong = 1uL,
    maxQueueSize: Int = 10,
    onEvent: (ULong) -> SafeFuture<ValidationResult> = { SafeFuture.completedFuture(ValidationResult.Companion.Valid) },
  ): TopicHandlerWithInOrderDelivering<ULong> {
    val subscriptionManager = SubscriptionManager<ULong>()
    subscriptionManager.subscribeToBlocks(onEvent)
    val deserializer =
      object : Deserializer<ULong> {
        override fun deserialize(bytes: ByteArray): ULong = Bytes.wrap(bytes).toLong().toULong()
      }

    val extractor = SequenceNumberExtractor<ULong> { it }
    return TopicHandlerWithInOrderDelivering(
      initialExpectedSequenceNumber = initialSeq,
      subscriptionManager = subscriptionManager,
      deserializer = deserializer,
      sequenceNumberExtractor = extractor,
      topicId = "test-topic",
      maxQueueSize = maxQueueSize,
    )
  }

  private fun makeMessage(seq: ULong): PreparedGossipMessage {
    val bytes = Bytes.ofUnsignedLong(seq.toLong())
    return MaruPreparedGossipMessage(
      origMessage = bytes,
      arrTimestamp = Optional.of(UInt64.valueOf(0)),
      domain = "test-domain",
      topicId = "test-topic",
    )
  }

  @Test
  fun `accepts and processes in-order message`() {
    val handler = makeHandler(initialSeq = 1uL)
    val msg1 = makeMessage(1uL)
    val result1 = handler.handleMessage(msg1).join()
    assertThat(result1).isEqualTo(Libp2pValidationResult.Valid)
    val msg2 = makeMessage(2uL)
    val result2 = handler.handleMessage(msg2).join()
    assertThat(result2).isEqualTo(Libp2pValidationResult.Valid)
  }

  @Test
  fun `queues out-of-order message and processes when expected`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler(initialSeq = 1uL) {
        processed.add(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    val msg2 = makeMessage(2uL)
    val msg1 = makeMessage(1uL)
    val future2 = handler.handleMessage(msg2)
    assertThat(future2.isDone).isFalse
    val future1 = handler.handleMessage(msg1)
    assertThat(future1.join()).isEqualTo(Libp2pValidationResult.Valid)
    assertThat(future2.join()).isEqualTo(Libp2pValidationResult.Valid)
    assertThat(processed).containsExactly(1uL, 2uL)
  }

  @Test
  fun `ignores duplicate or old messages`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler(initialSeq = 10uL) {
        processed.add(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    val msg1 = makeMessage(9uL)
    val result = handler.handleMessage(msg1).join()
    assertThat(result).isEqualTo(Libp2pValidationResult.Ignore)
    assertThat(processed).isEmpty()
  }

  @Test
  fun `ignores out-of-order message if queue is full`() {
    val handler = makeHandler(initialSeq = 1uL)
    // Fill the queue
    for (i in 2uL..11uL) {
      handler.handleMessage(makeMessage(i))
    }
    // This one should be ignored
    val result = handler.handleMessage(makeMessage(12uL)).join()
    assertThat(result).isEqualTo(Libp2pValidationResult.Ignore)
  }

  @Test
  fun `returns Ignore when deserializer throws exception`() {
    val subscriptionManager = SubscriptionManager<ULong>()
    subscriptionManager.subscribeToBlocks { SafeFuture.completedFuture(ValidationResult.Companion.Valid) }
    val deserializer =
      object : Deserializer<ULong> {
        override fun deserialize(bytes: ByteArray): ULong = throw IllegalArgumentException("bad data")
      }
    val extractor = SequenceNumberExtractor<ULong> { it }
    val handler =
      TopicHandlerWithInOrderDelivering(
        initialExpectedSequenceNumber = 1uL,
        subscriptionManager = subscriptionManager,
        deserializer = deserializer,
        sequenceNumberExtractor = extractor,
        topicId = "test-topic",
      )
    val msg = makeMessage(1uL)
    assertThat(handler.handleMessage(msg).join()).isEqualTo(Libp2pValidationResult.Invalid)
  }

  @Test
  fun `processed pending messages after the gap was filled`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler(initialSeq = 1uL) {
        processed.add(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    // Fill the queue
    val futures =
      (2uL..5uL)
        .map { msgNumber -> handler.handleMessage(makeMessage(msgNumber)) }
    // Now process the first message to drain the queue
    handler.handleMessage(makeMessage(1uL)).join()
    // Try to add another out-of-order message after draining
    val result2 = handler.handleMessage(makeMessage(6uL)).join()
    assertThat(result2).isEqualTo(Libp2pValidationResult.Valid)
    assertThat(processed).containsExactly(1uL, 2uL, 3uL, 4uL, 5uL, 6uL)
    assertThat(futures.map { it.get() }).allMatch { it == Libp2pValidationResult.Valid }
  }

  @Test
  fun `should be resilient to failures - subscriber throws`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler(initialSeq = 1uL) {
        processed.add(it)
        when {
          it.toInt() % 2 == 1 -> throw RuntimeException("Subscriber failure")
          else -> SafeFuture.completedFuture(ValidationResult.Companion.Valid)
        }
      }
    val futures =
      listOf(2uL, 4uL, 3uL, 5uL, 1uL)
        .map { seq -> seq to handler.handleMessage(makeMessage(seq)) }
        .map { (messageId, future) -> messageId to future.join() }
    assertThat(futures).isEqualTo(
      listOf(
        2uL to Libp2pValidationResult.Valid,
        4uL to Libp2pValidationResult.Valid,
        3uL to Libp2pValidationResult.Invalid,
        5uL to Libp2pValidationResult.Invalid,
        1uL to Libp2pValidationResult.Invalid,
      ),
    )
    assertThat(processed).containsExactly(1uL, 2uL, 3uL, 4uL, 5uL)
  }

  @Test
  fun `should be resilient to failures - subscriber rejects promise`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler(initialSeq = 1uL) {
        processed.add(it)
        when {
          it.toInt() % 2 == 1 -> SafeFuture.failedFuture(RuntimeException("Subscriber failure"))
          else -> SafeFuture.completedFuture(ValidationResult.Companion.Valid)
        }
      }
    // add 1ul at to test out of order message handling
    val futures =
      listOf(2uL, 4uL, 3uL, 5uL, 1uL)
        .map { seq -> seq to handler.handleMessage(makeMessage(seq)) }
        .map { (messageId, future) -> messageId to future.join() }

    assertThat(futures).isEqualTo(
      listOf(
        2uL to Libp2pValidationResult.Valid,
        4uL to Libp2pValidationResult.Valid,
        3uL to Libp2pValidationResult.Invalid,
        5uL to Libp2pValidationResult.Invalid,
        1uL to Libp2pValidationResult.Invalid,
      ),
    )
    assertThat(processed).containsExactly(1uL, 2uL, 3uL, 4uL, 5uL)
  }
}
