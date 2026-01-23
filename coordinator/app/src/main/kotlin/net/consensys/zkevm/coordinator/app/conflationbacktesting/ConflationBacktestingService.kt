package net.consensys.zkevm.coordinator.app.conflationbacktesting

import io.vertx.core.Vertx
import linea.coordinator.config.v2.CoordinatorConfig
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.metrics.MetricsFacade
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.ConcurrentMap

class ConflationBacktestingService(
  private val vertx: Vertx,
  private val configs: CoordinatorConfig,
  private val metricsFacade: MetricsFacade,
  private val httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory = VertxHttpJsonRpcClientFactory(
    vertx = vertx,
    metricsFacade = metricsFacade,
  ),
) {
  enum class ConflationBacktestingJobStatus {
    IN_PROGRESS,
    COMPLETED,
  }

  private val conflationBackTestingApps: ConcurrentMap<String, ConflationBacktestingApp> = ConcurrentHashMap()

  fun submitConflationBacktestingJob(conflationBacktestingConfig: ConflationBacktestingConfig): String {
    val jobId = conflationBacktestingConfig.jobId()
    val app = ConflationBacktestingApp(
      vertx = vertx,
      conflationBacktestingAppConfig = conflationBacktestingConfig,
      prodCoordinatorConfigs = configs,
      httpJsonRpcClientFactory = httpJsonRpcClientFactory,
      metricsFacade = metricsFacade,
    )
    conflationBackTestingApps[jobId] = app
    app.start()
    return jobId
  }

  fun getConflationBacktestingJobStatus(jobId: String): ConflationBacktestingJobStatus {
    return when (conflationBackTestingApps[jobId]?.isConflationBacktestingComplete()) {
      true -> ConflationBacktestingJobStatus.COMPLETED
      false -> ConflationBacktestingJobStatus.IN_PROGRESS
      null -> throw IllegalArgumentException("No job found with ID: $jobId")
    }
  }
}
