package net.consensys.zkevm.ethereum.coordination

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.logging.log4j.core.test.appender.ListAppender
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import java.util.concurrent.atomic.AtomicInteger
import java.util.function.Consumer

class EventDispatcherTest {

  private lateinit var logger: Logger

  @BeforeEach
  fun setup() {
    logger = LogManager.getLogger(EventDispatcher::class.java)
  }

  @Test
  fun dispatcherContinuesAfterFailedAttempt() {
    val number1 = AtomicInteger(0)
    val number2 = AtomicInteger(0)

    val mappedConsumers: Map<Consumer<AtomicInteger>, String> = mapOf(
      Consumer<AtomicInteger> { atomicInteger ->
        atomicInteger.incrementAndGet()
      } to "Consumer counter 1",
      Consumer<AtomicInteger> {
        throw RuntimeException("Test error")
      } to "Consumer error",
      Consumer<AtomicInteger> { atomicInteger ->
        atomicInteger.incrementAndGet()
      } to "Consumer counter 2"
    )

    val eventDispatcher = EventDispatcher(mappedConsumers)

    eventDispatcher.accept(number1)
    assertThat(number1.get()).isEqualTo(2)
    assertThat(number2.get()).isEqualTo(0)

    eventDispatcher.accept(number2)
    assertThat(number1.get()).isEqualTo(2)
    assertThat(number2.get()).isEqualTo(2)
  }

  @Test
  fun dispatcherErrorPrintsCorrectly() {
    val listAppender = ListAppender("ListAppender").apply {
      start()
    }
    (logger as org.apache.logging.log4j.core.Logger).addAppender(listAppender)

    val consumerName = "Mock Consumer"
    val exceptionMessage = "Test error"
    val consumerError: Consumer<Any> = Consumer {
      throw RuntimeException(exceptionMessage)
    }

    val mappedConsumers: Map<Consumer<Any>, String> = mapOf(
      consumerError to consumerName
    )

    val eventDispatcher = EventDispatcher(mappedConsumers)

    val event = { }
    eventDispatcher.accept(event)

    val logEvents = listAppender.events
    assertThat(logEvents).hasSize(1)
    val logEvent = logEvents[0]
    assertThat(logEvent.message.formattedMessage).contains("Failed to consume event")
    assertThat(logEvent.message.formattedMessage).contains(consumerName)
    assertThat(logEvent.message.formattedMessage).contains(event.toString())
    assertThat(logEvent.message.formattedMessage).contains(exceptionMessage)
  }
}
