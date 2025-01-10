package linea.staterecover.plugin

import io.micrometer.core.instrument.simple.SimpleMeterRegistry
import io.vertx.core.Vertx
import linea.staterecover.BlockHeaderStaticFields
import linea.staterecover.InMemoryRecoveryStatus
import linea.staterecover.RecoveryStatusPersistence
import linea.staterecover.StateRecoverApp
import linea.staterecover.clients.ExecutionLayerInProcessClient
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.plugin.BesuPlugin
import org.hyperledger.besu.plugin.ServiceManager
import org.hyperledger.besu.plugin.services.BesuEvents
import org.hyperledger.besu.plugin.services.BesuService
import org.hyperledger.besu.plugin.services.BlockSimulationService
import org.hyperledger.besu.plugin.services.BlockchainService
import org.hyperledger.besu.plugin.services.PicoCLIOptions
import org.hyperledger.besu.plugin.services.mining.MiningService
import org.hyperledger.besu.plugin.services.p2p.P2PService
import org.hyperledger.besu.plugin.services.sync.SynchronizationService

fun <T : BesuService> ServiceManager.getServiceOrThrow(clazz: Class<T>): T {
  return this.getService(clazz)
    .orElseThrow { IllegalStateException("${clazz.name} is not present in BesuContext") }
}

open class LineaStateRecoverPlugin : BesuPlugin {
  private val log: Logger = LogManager.getLogger(LineaStateRecoverPlugin::class.java)
  private val vertx = Vertx.vertx()
  private val cliOptions = PluginCliOptions()
  private lateinit var serviceManager: ServiceManager
  private lateinit var recoveryModeManager: RecoveryModeManager
  private lateinit var recoveryStatusPersistence: RecoveryStatusPersistence
  private lateinit var stateRecoverApp: StateRecoverApp

  override fun register(serviceManager: ServiceManager) {
    log.info("LineaStateRecoverPlugin Registering")
    this.serviceManager = serviceManager
    this.recoveryStatusPersistence = InMemoryRecoveryStatus()
    serviceManager
      .getServiceOrThrow(PicoCLIOptions::class.java)
      .addPicoCLIOptions(PluginCliOptions.cliOptionsPrefix, cliOptions)
    log.info("LineaStateRecoverPlugin Registered")
  }

  override fun start() {
    val config = cliOptions.getConfig()
    log.info("LineaStateRecoverPlugin starting: config={}", config)

    val blockchainService = serviceManager.getServiceOrThrow(BlockchainService::class.java)
    val synchronizationService = serviceManager.getServiceOrThrow(SynchronizationService::class.java)
    this.recoveryModeManager = RecoveryModeManager(
      p2pService = serviceManager.getServiceOrThrow(P2PService::class.java),
      miningService = serviceManager.getServiceOrThrow(MiningService::class.java),
      recoveryStatePersistence = this.recoveryStatusPersistence,
      synchronizationService = synchronizationService
    )
    val simulatorService = serviceManager.getServiceOrThrow(BlockSimulationService::class.java)
    val executionLayerClient = ExecutionLayerInProcessClient.create(
      blockchainService = blockchainService,
      stateRecoveryModeManager = this.recoveryModeManager,
      stateRecoveryStatusPersistence = this.recoveryStatusPersistence,
      simulatorService = simulatorService,
      synchronizationService = synchronizationService
    )
    this.stateRecoverApp = createAppAllInProcess(
      vertx = vertx,
      // FIXME: check latter on if metrics are exported
      meterRegistry = SimpleMeterRegistry(),
      elClient = executionLayerClient,
      stateManagerClientEndpoint = config.shomeiEndpoint,
      l1RpcEndpoint = config.l1RpcEndpoint,
      blobScanEndpoint = config.blobscanEndpoint,
      blockHeaderStaticFields = BlockHeaderStaticFields.localDev,
      appConfig = StateRecoverApp.Config(
        smartContractAddress = config.l1SmartContractAddress.toString(),
        l1LatestSearchBlock = net.consensys.linea.BlockParameter.Tag.LATEST,
        overridingRecoveryStartBlockNumber = config.overridingRecoveryStartBlockNumber
      )
    )
    serviceManager
      .getServiceOrThrow(BesuEvents::class.java)
      .addBlockAddedListener(recoveryModeManager)
    this.stateRecoverApp.start().get()
    log.info(
      "LineaStateRecoverPlugin started: recoveryStartBlockNumber={}",
      this.recoveryStatusPersistence.getRecoveryStartBlockNumber()
    )
  }

  override fun afterExternalServicePostMainLoop() {
    recoveryStatusPersistence.getRecoveryStartBlockNumber()?.let(recoveryModeManager::setTargetBlockNumber)
  }

  override fun stop() {
    // no-op
  }
}
