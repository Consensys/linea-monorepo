package linea

import io.vertx.core.Vertx
import linea.timer.TimerSchedule
import linea.timer.VertxTimerFactory
import org.apache.logging.log4j.Logger
import kotlin.time.Duration.Companion.milliseconds

abstract class PeriodicPollingService(
  private val vertx: Vertx,
  private val pollingIntervalMs: Long,
  private val log: Logger,
  private val name: String,
) : GenericPeriodicPollingService(
  timerFactory = VertxTimerFactory(vertx),
  timerSchedule = TimerSchedule.FIXED_DELAY,
  pollingInterval = pollingIntervalMs.milliseconds,
  log = log,
  name = name,
) {
  override fun handleError(error: Throwable) {
    log.error("Error with period polling service={}: errorMessage={}", name, error.message, error)
  }
}
