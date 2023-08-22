package net.consensys.linea.consensus.app

import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.Future
import io.vertx.core.Vertx
import io.vertx.core.VertxOptions
import io.vertx.micrometer.MicrometerMetricsOptions
import io.vertx.micrometer.VertxPrometheusOptions
import io.vertx.micrometer.backends.BackendRegistries
import net.consensys.linea.consensus.ForkChoicePoller
import net.consensys.linea.consensus.ForkChoiceStateClient
import net.consensys.linea.consensus.ForkChoiceStateJsonRpcClient
import net.consensys.linea.consensus.app.api.ObservabilityApi
import net.consensys.linea.forkchoicestate.ForkChoiceStateController
import net.consensys.linea.forkchoicestate.api.ForkChoiceStateApi
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.jwt.loadJwtSecretFromFile
import net.consensys.linea.vertx.ObservabilityServer
import net.consensys.linea.vertx.loadVertxConfig
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.ethereum.executionclient.ExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.auth.JwtConfig
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3JExecutionEngineClient
import tech.pegasys.teku.ethereum.executionclient.web3j.Web3jClientBuilder
import tech.pegasys.teku.infrastructure.time.SystemTimeProvider
import java.time.Duration
import java.util.Optional

class ConsensusLayerApp(config: ConsensusLayerAppConfig) {
  private val log = LogManager.getLogger(this::class.java)
  private val vertx: Vertx
  private val observabilityApi: ObservabilityApi
  private val meterRegistry: MeterRegistry
  private val forkChoiceStateClient: ForkChoiceStateClient
  private val forkChoiceStateApi: ForkChoiceStateApi?
  private val forkChoicePoller: ForkChoicePoller
  private val httpExecutionClient: Web3JClient =
    Web3jClientBuilder()
      .endpoint(config.executionClient.url.toString())
      .timeout(Duration.ofSeconds(20))
      .timeProvider(SystemTimeProvider())
      .executionClientEventsPublisher { false }
      .jwtConfigOpt(
        Optional.of(JwtConfig(loadJwtSecretFromFile(config.executionClient.jwtSecretFile)))
      )
      .build()
  private val executionClient: ExecutionEngineClient =
    Web3JExecutionEngineClient(httpExecutionClient)

  init {
    log.debug("System properties: {}", System.getProperties())
    val vertxConfigJson = loadVertxConfig(System.getProperty("vertx.configurationFile"))
    log.info("Vertx custom configs: {}", vertxConfigJson)
    val vertxConfig =
      VertxOptions(vertxConfigJson)
        .setMetricsOptions(
          MicrometerMetricsOptions()
            .setJvmMetricsEnabled(true)
            .setPrometheusOptions(
              VertxPrometheusOptions().setPublishQuantiles(true).setEnabled(true)
            )
            .setEnabled(true)
        )
    log.debug("Vertx full configs: {}", vertxConfig)
    log.info("App configs: {}", config)
    this.vertx = Vertx.vertx(vertxConfig)
    this.meterRegistry = BackendRegistries.getDefaultNow()
    this.observabilityApi =
      ObservabilityApi(
        ObservabilityServer.Config(
          "linea-consensus-client",
          config.api.observabilityPort.toInt()
        ),
        vertx
      )
    this.forkChoiceStateClient =
      ForkChoiceStateJsonRpcClient(
        VertxHttpJsonRpcClientFactory(vertx, meterRegistry)
          .create(config.forkChoiceSource.url, 2)
      )

    this.forkChoiceStateApi =
      config.api.forkChoiceProvider?.let {
        ForkChoiceStateApi(
          config.api.forkChoiceProvider,
          vertx,
          meterRegistry,
          ForkChoiceStateController(forkChoiceStateClient.getForkChoiceState().get())
        )
      }
    this.forkChoicePoller =
      ForkChoicePoller(
        vertx,
        config.forkChoiceSource.pollingInterval,
        forkChoiceStateClient,
        executionClient,
        false
      )

    // update client with latest update
    this.forkChoicePoller
      .updateExecutionLayer(this.forkChoiceStateClient.getForkChoiceState().get())
      .get()
  }

  fun start(): Future<*> {
    forkChoicePoller.startPoller()
    return (forkChoiceStateApi?.start() ?: Future.succeededFuture(null))
      .compose { observabilityApi.start() }
      .onComplete { log.info("App successfully started") }
  }

  fun stop(): Future<*> {
    log.info("Shooting down app..")
    return (forkChoiceStateApi?.start() ?: Future.succeededFuture(null))
      .compose { observabilityApi.stop() }
      .onComplete { log.info("App successfully closed") }
  }
}
