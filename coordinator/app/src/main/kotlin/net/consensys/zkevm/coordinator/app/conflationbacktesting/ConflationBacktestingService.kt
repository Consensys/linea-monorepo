package net.consensys.zkevm.coordinator.app.conflationbacktesting

import io.vertx.core.Vertx
import io.vertx.core.impl.ConcurrentHashSet
import linea.coordinator.config.v2.CoordinatorConfig
import linea.timer.TimerSchedule
import linea.timer.VertxPeriodicPollingService
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.metrics.MetricsFacade
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.ConcurrentHashMap
import kotlin.time.Duration.Companion.seconds

class ConflationBacktestingService(
  private val vertx: Vertx,
  private val configs: CoordinatorConfig,
  private val metricsFacade: MetricsFacade,
  private val httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory = VertxHttpJsonRpcClientFactory(
    vertx = vertx,
    metricsFacade = metricsFacade,
  ),
  private val log: Logger = LogManager.getLogger(ConflationBacktestingService::class.java),
) : VertxPeriodicPollingService(
  vertx = vertx,
  name = "ConflationBacktestingService",
  pollingIntervalMs = 1.seconds.inWholeMilliseconds,
  log = log,
  timerSchedule = TimerSchedule.FIXED_DELAY,
) {
  enum class ConflationBacktestingJobStatus {
    IN_PROGRESS,
    COMPLETED,
  }

  private val completedJobs: MutableSet<String> = ConcurrentHashSet()
  private val conflationBackTestingApps: MutableMap<String, ConflationBacktestingApp> = ConcurrentHashMap()

  fun submitConflationBacktestingJob(conflationBacktestingConfig: ConflationBacktestingConfig): String {
    val jobId = conflationBacktestingConfig.jobId()
    if (conflationBackTestingApps.containsKey(jobId) || completedJobs.contains(jobId)) {
      throw IllegalArgumentException("Conflation backtesting job with id $jobId already exists")
    }

    val app = ConflationBacktestingApp(
      vertx = vertx,
      conflationBacktestingAppConfig = conflationBacktestingConfig,
      prodCoordinatorConfigs = configs,
      httpJsonRpcClientFactory = httpJsonRpcClientFactory,
      metricsFacade = metricsFacade,
    )
    conflationBackTestingApps[jobId] = app
    app.start().thenPeek {
      log.info("Conflation backtesting job completed: jobId={}", jobId)
    }.exceptionally { error ->
      log.error("Conflation backtesting job failed: jobId={}, errorMessage={}", jobId, error.message, error)
    }
    return jobId
  }

  fun getConflationBacktestingJobStatus(jobId: String): ConflationBacktestingJobStatus {
    if (completedJobs.contains(jobId)) {
      return ConflationBacktestingJobStatus.COMPLETED
    }
    if (conflationBackTestingApps.containsKey(jobId)) {
      return ConflationBacktestingJobStatus.IN_PROGRESS
    }
    throw IllegalArgumentException("No conflation backtesting job found with id: $jobId")
  }

  override fun action(): SafeFuture<*> {
    val completedJobIds = mutableListOf<String>()
    conflationBackTestingApps.forEach { (jobId, app) ->
      if (app.isConflationBacktestingComplete()) {
        completedJobIds.add(jobId)
      }
    }
    completedJobIds.forEach { jobId ->
      completedJobs.add(jobId)
      conflationBackTestingApps.remove(jobId)
    }
    return SafeFuture.completedFuture(Unit)
  }
}
