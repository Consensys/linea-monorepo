package net.consensys.zkevm.coordinator.app

import com.fasterxml.jackson.databind.module.SimpleModule
import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.Vertx
import io.vertx.core.VertxOptions
import io.vertx.core.json.jackson.DatabindCodec
import io.vertx.micrometer.MicrometerMetricsOptions
import io.vertx.micrometer.VertxPrometheusOptions
import io.vertx.micrometer.backends.BackendRegistries
import io.vertx.sqlclient.SqlClient
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.jsonrpc.client.LoadBalancingJsonRpcClient
import net.consensys.linea.jsonrpc.client.VertxHttpJsonRpcClientFactory
import net.consensys.linea.vertx.loadVertxConfig
import net.consensys.zkevm.coordinator.api.Api
import net.consensys.zkevm.coordinator.blockcreation.BlockCreationMonitor
import net.consensys.zkevm.coordinator.blockcreation.ExtendedWeb3JImpl
import net.consensys.zkevm.coordinator.blockcreation.GethCliqueSafeBlockProvider
import net.consensys.zkevm.coordinator.blockcreation.TracesFilesManager
import net.consensys.zkevm.coordinator.clients.FileBasedProverClient
import net.consensys.zkevm.coordinator.clients.TracesGeneratorJsonRpcClientV1
import net.consensys.zkevm.coordinator.clients.Type2StateManagerClient
import net.consensys.zkevm.coordinator.clients.Type2StateManagerJsonRpcClient
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.coordination.conflation.BlockToBatchSubmissionCoordinator
import net.consensys.zkevm.ethereum.coordination.conflation.BlocksTracesConflationCalculatorImpl
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationCalculatorConfig
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationService
import net.consensys.zkevm.ethereum.coordination.conflation.ConflationServiceImpl
import net.consensys.zkevm.ethereum.coordination.conflation.DataLimits
import net.consensys.zkevm.ethereum.coordination.conflation.TracesConflationCoordinatorImpl
import net.consensys.zkevm.ethereum.coordination.proofcreation.FileBasedZkProofCreationCoordinator
import net.consensys.zkevm.ethereum.finalization.FinalizationMonitor
import net.consensys.zkevm.ethereum.settlement.persistence.Db
import net.consensys.zkevm.ethereum.settlement.persistence.PostgresBatchesRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.utils.Async
import tech.pegasys.teku.ethereum.executionclient.serialization.BytesSerializer
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.net.URL
import kotlin.time.toKotlinDuration

class CoordinatorApp(private val configs: CoordinatorConfig) {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val vertx: Vertx = run {
    log.trace("System properties: {}", System.getProperties())
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
    log.info("App configs: {}", configs)

    // TODO: adapt JsonMessageProcessor to use custom ObjectMapper
    // this is just dark magic.
    val module = SimpleModule()
    module.addSerializer(Bytes::class.java, BytesSerializer())
    DatabindCodec.mapper().registerModule(module)
    // .enable(SerializationFeature.INDENT_OUTPUT)
    Vertx.vertx(vertxConfig)
  }
  private val meterRegistry: MeterRegistry = BackendRegistries.getDefaultNow()
  private val httpJsonRpcClientFactory = VertxHttpJsonRpcClientFactory(vertx, meterRegistry)
  private val api = Api(
    Api.Config(
      configs.api.observabilityPort
    ),
    vertx
  )
  private val l2Web3jClient: Web3j =
    Web3j.build(
      HttpService(configs.sequencer.ethApi.toString()),
      1000,
      Async.defaultExecutorService()
    )
  private val l2ZkGethWeb3jClient: Web3j =
    Web3j.build(
      HttpService(configs.zkGethTraces.ethApi.toString()),
      1000,
      Async.defaultExecutorService()
    )
  private val proverClient: FileBasedProverClient =
    FileBasedProverClient(
      config =
      FileBasedProverClient.Config(
        requestDirectory = configs.prover.fsInputDirectory,
        responseDirectory = configs.prover.fsOutputDirectory,
        inprogessProvingSuffixPattern = configs.prover.fsInprogessProvingSuffixPattern,
        pollingInterval = configs.prover.fsPollingInterval.toKotlinDuration(),
        timeout = configs.prover.timeout.toKotlinDuration(),
        proverVersion = configs.prover.version,
        l2MessageServiceAddress = configs.l2.messageServiceAddress,
        tracesVersion = configs.traces.version,
        stateManagerVersion = configs.stateManager.version
      ),
      vertx = vertx,
      l2Web3jClient = l2Web3jClient
    )

  private val sqlClient: SqlClient = initDb(configs.database)
  private val batchesRepository =
    PostgresBatchesRepository(
      PostgresBatchesRepository.Config(configs.batchSubmission.maxBatchesToSendPerTick.toUInt()),
      sqlClient,
      proverClient.proverResponsesRepository
    )
  private val l1App = L1DependentApp(
    configs,
    vertx,
    l2Web3jClient,
    httpJsonRpcClientFactory,
    batchesRepository
  ) { update: FinalizationMonitor.FinalizationUpdate ->
    batchesRepository.setBatchStatusUpToEndBlockNumber(
      update.blockNumber,
      Batch.Status.Pending,
      Batch.Status.Finalized
    )
  }

  private val lastFinalizedBlockNumber: ULong = run {
    when {
      configs.conflation.forceStartingBlock != null -> {
        configs.conflation.forceStartingBlock.toULong() - 1UL
      }

      !configs.testL1Disabled -> l1App.lastFinalizedBlock().get()
      // Beginning of Rollup
      else -> 0UL
    }.also {
      log.info("Last finalized block: {}", it)
    }
  }

  private val extendedWeb3j = ExtendedWeb3JImpl(l2ZkGethWeb3jClient)
  private val blockCreationMonitor = run {
    log.info("Resuming conflation from block {}", lastFinalizedBlockNumber + 1UL)
    val parentBlock = extendedWeb3j.ethGetExecutionPayloadByNumber(lastFinalizedBlockNumber.toLong()).get()
    BlockCreationMonitor(
      vertx,
      extendedWeb3j,
      parentBlock.blockNumber.longValue() + 1,
      parentBlock.blockHash,
      { blockEvent: BlockCreated -> block2BatchCoordinator.acceptBlock(blockEvent) },
      BlockCreationMonitor.Config(
        configs.zkGethTraces.newBlockPollingInterval.toKotlinDuration(),
        configs.l2.blocksToFinalization.toLong()
      )
    )
  }

  private val conflationService: ConflationService = ConflationServiceImpl(
    BlocksTracesConflationCalculatorImpl(
      lastFinalizedBlockNumber,
      GethCliqueSafeBlockProvider(
        extendedWeb3j,
        GethCliqueSafeBlockProvider.Config(configs.l2.blocksToFinalization.toLong())
      ),
      ConflationCalculatorConfig(
        tracesConflationLimit = configs.conflation.tracesLimits,
        dataConflationLimits = DataLimits(
          totalLimitBytes = configs.conflation.totalLimitBytes.toUInt(),
          perBlockOverheadBytes = configs.conflation.perBlockOverheadBytes.toUInt(),
          minBlockL1SizeBytes = configs.conflation.minBlockL1SizeBytes.toUInt()
        ),
        conflationDeadline = configs.conflation.conflationDeadline.toKotlinDuration(),
        conflationDeadlineCheckInterval = configs.conflation.conflationDeadlineCheckInterval.toKotlinDuration(),
        conflationDeadlineLastBlockConfirmationDelay =
        configs.conflation.conflationDeadlineLastBlockConfirmationDelay.toKotlinDuration(),
        blocksLimit = configs.conflation.blocksLimit?.toUInt()
      )
    )
  )

  private val block2BatchCoordinator = run {
    val tracesFileManager =
      TracesFilesManager(
        vertx,
        TracesFilesManager.Config(
          configs.traces.fileManager.rawTracesDirectory,
          configs.traces.fileManager.nonCanonicalRawTracesDirectory,
          configs.traces.fileManager.pollingInterval.toKotlinDuration(),
          configs.traces.fileManager.tracesFileCreationWaitTimeout.toKotlinDuration(),
          configs.traces.version,
          configs.traces.fileManager.tracesFileExtension,
          configs.traces.fileManager.createNonCanonicalDirectory
        )
      )
    val zkStateClient: Type2StateManagerClient = Type2StateManagerJsonRpcClient(
      LoadBalancingJsonRpcClient(
        configs.stateManager.endpoints.map { httpJsonRpcClientFactory.create(it) },
        configs.stateManager.requestLimitPerEndpoint
      ),
      Type2StateManagerJsonRpcClient.Config(configs.stateManager.version)
    )

    val tracesCountersClient = TracesGeneratorJsonRpcClientV1(
      vertx,
      createLoadBalancerRpcClient(
        httpJsonRpcClientFactory,
        configs.traces.counters.endpoints,
        configs.traces.counters.requestLimitPerEndpoint
      ),
      TracesGeneratorJsonRpcClientV1.Config(
        configs.traces.counters.requestMaxRetries.toInt(),
        configs.traces.counters.requestRetryInterval.toKotlinDuration()
      )
    )
    val tracesConflationClient = TracesGeneratorJsonRpcClientV1(
      vertx,
      createLoadBalancerRpcClient(
        httpJsonRpcClientFactory,
        configs.traces.conflation.endpoints,
        configs.traces.conflation.requestLimitPerEndpoint
      ),
      TracesGeneratorJsonRpcClientV1.Config(
        configs.traces.conflation.requestMaxRetries.toInt(),
        configs.traces.conflation.requestRetryInterval.toKotlinDuration()
      )
    )
    BlockToBatchSubmissionCoordinator(
      conflationService,
      tracesFileManager,
      tracesCountersClient,
      TracesConflationCoordinatorImpl(tracesConflationClient, zkStateClient),
      FileBasedZkProofCreationCoordinator(proverClient),
      l1App.batchSubmissionCoordinator,
      vertx
    )
  }

  init {
    log.info("Coordinator app instantiated")
  }

  fun start() {
    SafeFuture.allOf(
      l1App.start(),
      blockCreationMonitor.start()
    ).thenCompose {
      api.start().toSafeFuture()
    }.get()

    log.info("Started :)")
  }

  fun stop(): Int {
    SafeFuture.allOf(
      l1App.stop(),
      blockCreationMonitor.stop(),
      SafeFuture.fromRunnable { l2Web3jClient.shutdown() },
      api.stop().toSafeFuture()
    ).thenCompose {
      vertx.close().toSafeFuture()
    }
      .get()
    return 0
  }

  data class BlockNumberAndHash(
    val blockNumber: ULong,
    val blockHash: Bytes32,
    val parentHash: Bytes32
  )

  private fun initDb(dbConfig: DatabaseConfig): SqlClient {
    Db.applyDbMigrations(
      dbConfig.host,
      dbConfig.port,
      dbConfig.schema,
      dbConfig.username,
      dbConfig.password.value
    )
    return Db.vertxSqlClient(
      vertx,
      dbConfig.host,
      dbConfig.port,
      dbConfig.schema,
      dbConfig.username,
      dbConfig.password.value,
      dbConfig.transactionalPoolSize,
      dbConfig.readPipeliningLimit
    )
  }

  private fun createLoadBalancerRpcClient(
    httpJsonRpcClientFactory: VertxHttpJsonRpcClientFactory,
    endpoints: List<URL>,
    requestLimitPerEndpoint: UInt
  ): LoadBalancingJsonRpcClient {
    return LoadBalancingJsonRpcClient(
      endpoints.map(httpJsonRpcClientFactory::create),
      requestLimitPerEndpoint
    )
  }
}
