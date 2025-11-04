package net.consensys.zkevm

import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration

abstract class LineaPeriodicPollingService(
  private val timerFactory: TimerFactory,
  private val pollingInterval: Duration,
  private val log: Logger,
) : LongRunningService {
  init {
    require(pollingInterval.inWholeMilliseconds > 0) {
      "pollingInterval must be greater than 0"
    }
  }
  private var timer: LineaTimer? = null

  abstract fun action(): SafeFuture<*>

  open fun handleError(error: Throwable) {
    log.warn("Error with periodic polling service: errorMessage={}", error.message, error)
  }

  @Synchronized
  override fun start(): SafeFuture<Unit> {
    if (timer == null) {
      timer = timerFactory.createTimer(
        name = this.javaClass.simpleName,
        task = { action().get() },
        initialDelay = pollingInterval,
        period = pollingInterval,
        errorHandler = this::handleError,
      )
      timer!!.start()
    }
    return SafeFuture.completedFuture(Unit)
  }

  @Synchronized
  override fun stop(): SafeFuture<Unit> {
    if (timer != null) {
      timer!!.stop()
      timer = null
    } else {
      log.info("Service is not running to stop it!")
    }
    return SafeFuture.completedFuture(Unit)
  }
}
