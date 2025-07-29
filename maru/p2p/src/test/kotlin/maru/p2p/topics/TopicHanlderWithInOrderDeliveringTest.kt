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
import java.util.concurrent.TimeUnit
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
  var nextExpectedSequenceNumber = 1UL

  private fun makeHandler(
    maxQueueSize: Int = 10,
    onEvent: (ULong) -> SafeFuture<ValidationResult> = {
      nextExpectedSequenceNumber++
      SafeFuture.completedFuture(ValidationResult.Companion.Valid)
    },
  ): TopicHandlerWithInOrderDelivering<ULong> {
    val subscriptionManager = SubscriptionManager<ULong>()
    subscriptionManager.subscribeToBlocks(onEvent)
    val deserializer =
      object : Deserializer<ULong> {
        override fun deserialize(bytes: ByteArray): ULong = Bytes.wrap(bytes).toLong().toULong()
      }
    val nextExpectedSequenceNumberProvider: () -> ULong = { nextExpectedSequenceNumber }
    val extractor = SequenceNumberExtractor<ULong> { it }
    return TopicHandlerWithInOrderDelivering(
      subscriptionManager = subscriptionManager,
      deserializer = deserializer,
      sequenceNumberExtractor = extractor,
      topicId = "test-topic",
      maxQueueSize = maxQueueSize,
      nextExpectedSequenceNumberProvider = nextExpectedSequenceNumberProvider,
      isHandlingEnabled = { true },
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
    val handler = makeHandler()
    val msg1 = makeMessage(1uL)
    val result1 = handler.handleMessage(msg1).get(5, TimeUnit.SECONDS)
    assertThat(result1).isEqualTo(Libp2pValidationResult.Valid)
    val msg2 = makeMessage(2uL)
    val result2 = handler.handleMessage(msg2).get(5, TimeUnit.SECONDS)
    assertThat(result2).isEqualTo(Libp2pValidationResult.Valid)
  }

  @Test
  fun `queues out-of-order message and processes when expected`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler {
        processed.add(it)
        nextExpectedSequenceNumber++
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    val msg2 = makeMessage(2uL)
    val msg1 = makeMessage(1uL)
    val future2 = handler.handleMessage(msg2)
    assertThat(future2.isDone).isFalse
    val future1 = handler.handleMessage(msg1)
    assertThat(future1.get(5, TimeUnit.SECONDS)).isEqualTo(Libp2pValidationResult.Valid)
    assertThat(future2.get(5, TimeUnit.SECONDS)).isEqualTo(Libp2pValidationResult.Valid)
    assertThat(processed).containsExactly(1uL, 2uL)
  }

  @Test
  fun `ignores duplicate or old messages`() {
    val processed = mutableListOf<ULong>()
    nextExpectedSequenceNumber = 10UL
    val handler =
      makeHandler {
        nextExpectedSequenceNumber++
        processed.add(it)
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    val msg1 = makeMessage(9uL)
    val result = handler.handleMessage(msg1).get(5, TimeUnit.SECONDS)
    assertThat(result).isEqualTo(Libp2pValidationResult.Ignore)
    assertThat(processed).isEmpty()
  }

  @Test
  fun `ignores out-of-order message if queue is full`() {
    val handler = makeHandler()
    // Fill the queue
    for (i in 2uL..11uL) {
      handler.handleMessage(makeMessage(i))
    }
    // This one should be ignored
    val result = handler.handleMessage(makeMessage(12uL)).get(5, TimeUnit.SECONDS)
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
        nextExpectedSequenceNumberProvider = { 1uL },
        subscriptionManager = subscriptionManager,
        deserializer = deserializer,
        sequenceNumberExtractor = extractor,
        topicId = "test-topic",
        isHandlingEnabled = { true },
      )
    val msg = makeMessage(1uL)
    assertThat(handler.handleMessage(msg).get(5, TimeUnit.SECONDS)).isEqualTo(Libp2pValidationResult.Invalid)
  }

  @Test
  fun `processed pending messages after the gap was filled`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler {
        processed.add(it)
        nextExpectedSequenceNumber++
        SafeFuture.completedFuture(ValidationResult.Companion.Valid)
      }
    // Fill the queue
    val futures =
      (2uL..5uL)
        .map { msgNumber -> handler.handleMessage(makeMessage(msgNumber)) }
    // Now process the first message to drain the queue
    handler.handleMessage(makeMessage(1uL)).get(5, TimeUnit.SECONDS)
    // Try to add another out-of-order message after draining
    val result2 = handler.handleMessage(makeMessage(6uL)).get(5, TimeUnit.SECONDS)
    assertThat(result2).isEqualTo(Libp2pValidationResult.Valid)
    assertThat(processed).containsExactly(1uL, 2uL, 3uL, 4uL, 5uL, 6uL)
    assertThat(futures.map { it.get() }).allMatch { it == Libp2pValidationResult.Valid }
  }

  @Test
  fun `should be resilient to failures - subscriber throws`() {
    val processed = mutableListOf<ULong>()
    val handler =
      makeHandler {
        processed.add(it)
        val result =
          when {
            nextExpectedSequenceNumber.toInt() % 2 == 1 ->
              SafeFuture.failedFuture(
                RuntimeException("Subscriber failure"),
              )

            else -> {
              SafeFuture.completedFuture(ValidationResult.Companion.Valid)
            }
          }
        nextExpectedSequenceNumber++
        result as SafeFuture<ValidationResult>
      }
    val futures =
      listOf(2uL, 4uL, 3uL, 5uL, 1uL)
        .map { seq -> seq to handler.handleMessage(makeMessage(seq)) }
        .map { (messageId, future) -> messageId to future.get(5, TimeUnit.SECONDS) }
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
      makeHandler {
        processed.add(it)
        val result =
          when {
            nextExpectedSequenceNumber.toInt() % 2 == 1 ->
              SafeFuture.failedFuture(
                RuntimeException("Subscriber failure"),
              )

            else -> {
              SafeFuture.completedFuture(ValidationResult.Companion.Valid)
            }
          }
        nextExpectedSequenceNumber++
        result as SafeFuture<ValidationResult>
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
