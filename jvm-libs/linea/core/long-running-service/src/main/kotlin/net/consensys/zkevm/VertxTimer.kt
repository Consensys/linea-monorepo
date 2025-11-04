package net.consensys.zkevm

import io.vertx.core.Vertx
import java.util.UUID
import java.util.concurrent.Callable
import kotlin.time.Duration

class VertxTimer(
  private val vertx: Vertx,
  override val name: String,
  override val task: Runnable,
  override val initialDelay: Duration,
  override val period: Duration,
  override val errorHandler: (Throwable) -> Unit,
) : LineaTimer {
  private var timerId: Long? = null

  internal fun timerReference(): Long? = timerId

  @Synchronized
  override fun start() {
    if (timerId == null) {
      timerId = vertx.setTimer(initialDelay.inWholeMilliseconds, this::taskHandler)
    }
  }

  @Suppress("UNUSED_PARAMETER")
  private fun taskHandler(_timerId: Long) {
    val callable = Callable {
      task.run()
    }
    vertx.executeBlocking(callable).onComplete { result ->
      if (result.cause() != null) {
        errorHandler(result.cause())
      }
      synchronized(this) {
        if (timerId != null) {
          timerId = vertx.setTimer(period.inWholeMilliseconds, this::taskHandler)
        }
      }
    }
  }

  @Synchronized
  override fun stop() {
    if (timerId != null) {
      vertx.cancelTimer(timerId!!)
      timerId = null
    }
  }
}

class VertxTimerFactory(private val vertx: Vertx) : TimerFactory {
  override fun createTimer(
    name: String,
    task: Runnable,
    initialDelay: Duration,
    period: Duration,
    errorHandler: (Throwable) -> Unit,
  ): LineaTimer {
    return VertxTimer(
      vertx = vertx,
      name = "$name-${UUID.randomUUID()}",
      task = task,
      initialDelay = initialDelay,
      period = period,
      errorHandler = errorHandler,
    )
  }
}
