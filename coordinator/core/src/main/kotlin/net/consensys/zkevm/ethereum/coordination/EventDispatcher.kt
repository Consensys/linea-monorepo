package net.consensys.zkevm.ethereum.coordination

import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import java.util.function.Consumer

class EventDispatcher<T>(
  private val consumers: Map<Consumer<T>, String>
) : Consumer<T> {
  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun accept(event: T) {
    consumers.forEach { (consumer, name) ->
      try {
        consumer.accept(event)
      } catch (e: Exception) {
        log.warn(
          "Failed to consume event: consumer={} event={} errorMessage={}",
          name,
          event,
          e.message,
          e
        )
      }
    }
  }
}
