package net.consensys.zkevm

import java.util.Timer
import java.util.UUID
import kotlin.concurrent.timerTask
import kotlin.time.Duration

class JvmTimer(
  override val name: String,
  override val task: Runnable,
  override val initialDelay: Duration,
  override val period: Duration,
  override val errorHandler: (Throwable) -> Unit,
) : LineaTimer {
  private var timer: Timer? = null

  internal fun timerReference(): Timer? = timer

  @Synchronized
  override fun start() {
    if (timer != null) {
      return
    }
    timer = Timer(name, true)
    timer!!.scheduleAtFixedRate(
      timerTask {
        try {
          task.run()
        } catch (e: Throwable) {
          errorHandler(e)
        }
      },
      initialDelay.inWholeMilliseconds,
      period.inWholeMilliseconds,
    )
  }

  @Synchronized
  override fun stop() {
    if (timer == null) {
      return
    }
    timer?.cancel()
    timer = null
  }
}

class JvmTimerFactory : TimerFactory {
  override fun createTimer(
    name: String,
    task: Runnable,
    initialDelay: Duration,
    period: Duration,
    errorHandler: (Throwable) -> Unit,
  ): LineaTimer {
    return JvmTimer(
      name = "$name-${UUID.randomUUID()}",
      task = task,
      initialDelay = initialDelay,
      period = period,
      errorHandler = errorHandler,
    )
  }
}
