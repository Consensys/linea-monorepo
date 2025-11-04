package net.consensys.zkevm

import io.vertx.core.Vertx
import org.apache.logging.log4j.Logger
import kotlin.time.Duration.Companion.milliseconds

abstract class PeriodicPollingService(
  private val vertx: Vertx,
  private val pollingIntervalMs: Long,
  private val log: Logger,
) : LineaPeriodicPollingService(
  timerFactory = VertxTimerFactory(vertx),
  pollingInterval = pollingIntervalMs.milliseconds,
  log = log,
) {
  override fun handleError(error: Throwable) {
    log.error("Error with period polling service: errorMessage={}", error.message, error)
  }
}
